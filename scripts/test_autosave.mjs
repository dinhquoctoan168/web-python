import test from 'node:test';
import assert from 'node:assert/strict';
import ideWorkerModule from '../web/static/js/ide_worker.js';

const { createSaveManager } = ideWorkerModule;

// Helper tạo deferred promise để điều khiển thứ tự response chính xác
function createDeferred() {
    let resolve, reject;
    const promise = new Promise((res, rej) => {
        resolve = res;
        reject = rej;
    });
    return { promise, resolve, reject };
}

test('1. Worker tuần tự: V1 đang chạy, V2/V3 chờ -> Chỉ một save request chạy và bản cuối là V3', async () => {
    const serverLogs = [];
    const v1Deferred = createDeferred();

    let editorVal = 'print("hello")';
    let currentEx = { id: 1, initialCode: 'print("hello")' };

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        getCurrentExercise: () => currentEx,
        debounceMs: 20,
        apiSave: async (exerciseId, code) => {
            serverLogs.push({ event: 'start', code });
            if (code === 'V1') {
                await v1Deferred.promise;
            } else {
                await new Promise(r => setTimeout(r, 10));
            }
            serverLogs.push({ event: 'finish', code });
            return { ok: true, status: 200 };
        }
    });

    // 1. Gõ V1 và lưu ngay
    editorVal = 'V1';
    saveManager.queueDraftSave(1, 'V1', true);
    await new Promise(r => setTimeout(r, 5));

    // 2. Trong lúc V1 đang bay, gõ V2 rồi V3
    editorVal = 'V2';
    saveManager.queueDraftSave(1, 'V2', true);
    editorVal = 'V3';
    saveManager.queueDraftSave(1, 'V3', true);

    // 3. Cho V1 hoàn thành
    v1Deferred.resolve();

    const success = await saveManager.flushCurrentDraft();

    assert.equal(success, true, 'Flush phải thành công');
    assert.equal(saveManager.maxConcurrentRequests, 1, 'Số request đồng thời tối đa phải là 1 (Single-flight)');
    assert.equal(saveManager.getLastSavedCode(1), 'V3', 'Bản ghi cuối cùng được xác nhận phải là V3');
    assert.equal(saveManager.hasUnsavedChanges(), false, 'Không còn thay đổi chưa lưu');

    const finishedCodes = serverLogs.filter(l => l.event === 'finish').map(l => l.code);
    assert.deepEqual(finishedCodes, ['V1', 'V3'], 'Server chỉ nhận V1 rồi đến V3 (V2 đã được gộp)');
});

test('2. Đổi bài tải chậm: Flush A xong; giữ response B -> Editor vẫn khóa; reset/chọn bài khác bị chặn', async () => {
    let editorVal = 'code_A_edited';
    let readOnlyState = false;
    let currentEx = { id: 1, initialCode: 'code_A_initial' };

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx,
        apiSave: async () => ({ ok: true, status: 200 })
    });

    // Mô phỏng logic chuyển bài của IDE
    const bDeferred = createDeferred();

    async function switchExercise(targetId) {
        if (saveManager.isTransitionInProgress()) return false;
        if (!saveManager.startTransition()) return false;

        try {
            if (saveManager.hasUnconfirmedDrafts()) {
                const saved = await saveManager.flushCurrentDraft();
                if (!saved || saveManager.hasUnconfirmedDrafts()) return false;
            }

            // Đang fetch bài mới (bị delay)
            await bDeferred.promise;

            currentEx = { id: targetId, initialCode: 'code_B_initial' };
            editorVal = currentEx.initialCode;
            saveManager.setExercise(currentEx);
            return true;
        } finally {
            saveManager.endTransition();
        }
    }

    // Bắt đầu chuyển sang bài 2
    const switchPromise = switchExercise(2);
    await new Promise(r => setTimeout(r, 10));

    // Trong lúc bài 2 đang tải:
    assert.equal(saveManager.isTransitionInProgress(), true, 'Transition lock phải đang bật');
    assert.equal(readOnlyState, true, 'Editor phải đang ở chế độ readOnly');

    // Thao tác reset hoặc gõ phím bị chặn
    function tryResetCode() {
        if (saveManager.isTransitionInProgress()) return false;
        editorVal = currentEx.initialCode;
        return true;
    }
    assert.equal(tryResetCode(), false, 'Reset code phải bị chặn khi transition đang chạy');

    // Thao tác chuyển sang bài 3 cạnh tranh bị chặn
    const switch3Result = await switchExercise(3);
    assert.equal(switch3Result, false, 'Chuyển sang bài khác phải bị chặn khi transition đang chạy');

    // Cho request bài 2 hoàn thành
    bDeferred.resolve();
    const switchResult = await switchPromise;

    assert.equal(switchResult, true, 'Chuyển bài 2 thành công sau khi response');
    assert.equal(saveManager.isTransitionInProgress(), false, 'Transition lock phải được giải phóng');
    assert.equal(readOnlyState, false, 'Editor phải được mở khóa');
    assert.equal(currentEx.id, 2, 'Context hiện tại là bài 2');
});

test('3. Tải bài mới thất bại: API đọc lỗi -> Giữ bài cũ, mở khóa, không đổi context', async () => {
    let editorVal = 'code_A_important';
    let readOnlyState = false;
    let currentEx = { id: 1, initialCode: 'code_A_initial' };

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx,
        apiSave: async () => ({ ok: true, status: 200 })
    });

    let currentBreadcrumb = 'Bài 1';

    async function switchExerciseWithError(targetId) {
        if (saveManager.isTransitionInProgress()) return false;
        if (!saveManager.startTransition()) return false;

        try {
            if (saveManager.hasUnconfirmedDrafts()) {
                const saved = await saveManager.flushCurrentDraft();
                if (!saved || saveManager.hasUnconfirmedDrafts()) return false;
            }

            // Mô phỏng fetch bài mới thất bại (500)
            throw new Error('HTTP 500 Network Error');
        } catch (err) {
            return false;
        } finally {
            saveManager.endTransition();
        }
    }

    const res = await switchExerciseWithError(2);
    assert.equal(res, false, 'Chuyển bài thất bại');
    assert.equal(currentEx.id, 1, 'Vẫn giữ context bài cũ (id: 1)');
    assert.equal(editorVal, 'code_A_important', 'Code bài cũ còn nguyên');
    assert.equal(currentBreadcrumb, 'Bài 1', 'Breadcrumb không bị đổi sang bài lỗi');
    assert.equal(saveManager.isTransitionInProgress(), false, 'Transition lock phải được mở lại');
    assert.equal(readOnlyState, false, 'Editor được mở lại cho người dùng');
});

test('4. Khởi tạo nháp chậm: Khóa được giữ cho đến khi nạp draft bài mới xong -> Không ghi đè code mới', async () => {
    let editorVal = '';
    let readOnlyState = false;
    let currentEx = { id: 1, initialCode: 'code_A' };

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx,
        apiSave: async () => ({ ok: true, status: 200 })
    });

    const draftDeferred = createDeferred();

    async function switchAndLoadDraft(targetEx) {
        saveManager.startTransition();
        try {
            // Render starter code
            currentEx = targetEx;
            editorVal = targetEx.initialCode;
            saveManager.setLastSavedCode(targetEx.id, targetEx.initialCode);

            // Chờ draft nạp xong TRƯỚC KHI mở khóa
            const draft = await draftDeferred.promise;
            if (draft) {
                editorVal = draft;
                saveManager.setLastSavedCode(targetEx.id, draft);
            }
        } finally {
            saveManager.endTransition();
        }
    }

    const switchPromise = switchAndLoadDraft({ id: 2, initialCode: 'starter_B' });
    await new Promise(r => setTimeout(r, 5));

    // Trong lúc draft chưa về, editor vẫn bị khóa nên người dùng không gõ đè
    assert.equal(readOnlyState, true, 'Editor vẫn khóa trong lúc draft đang nạp');

    // Draft server trả về
    draftDeferred.resolve('saved_draft_B');
    await switchPromise;

    assert.equal(editorVal, 'saved_draft_B', 'Draft hợp lệ được nạp thành công');
    assert.equal(readOnlyState, false, 'Chỉ mở khóa sau khi draft đã hoàn tất');
});

test('5. Pending của bài cũ: Editor bài hiện tại sạch, worker còn snapshot A -> Home chờ lưu A', async () => {
    let editorVal = 'starter_B';
    let currentEx = { id: 2, initialCode: 'starter_B' }; // Bài 2 sạch 100%
    const saveDeferredA = createDeferred();
    let savedA = false;

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        getCurrentExercise: () => currentEx,
        apiSave: async (exId, code) => {
            if (exId === 1) {
                await saveDeferredA.promise;
                savedA = true;
            }
            return { ok: true, status: 200 };
        }
    });

    // Giả lập worker còn snapshot bài 1 đang chờ lưu
    saveManager.queueDraftSave(1, 'code_A_snapshot', true);

    // Bài 2 hiện tại hoàn toàn sạch
    assert.equal(saveManager.hasUnsavedChanges(), false, 'Bài 2 không có thay đổi chưa lưu trong editor');
    assert.equal(saveManager.hasUnconfirmedDrafts(), true, 'Nhưng hệ thống vẫn còn snapshot chưa lưu của bài 1');

    let navigated = false;
    const navPromise = saveManager.handleSafeNavigation('/home', null);

    // handleSafeNavigation không được rời ngay
    await new Promise(r => setTimeout(r, 10));
    assert.equal(savedA, false, 'Chưa flush xong bài 1');

    // Cho bài 1 lưu xong
    saveDeferredA.resolve();
    const navResult = await navPromise;

    assert.equal(navResult, true, 'Điều hướng thành công sau khi bài 1 đã lưu xong');
    assert.equal(savedA, true, 'Bài 1 đã được xác nhận lưu trên máy chủ');
    assert.equal(saveManager.hasUnconfirmedDrafts(), false, 'Mọi draft đã lưu sạch 100%');
});

test('6. Code quay về bản đã lưu: Request code khác đang chạy, editor trở lại cũ -> Bản cuối khớp editor trước khi rời', async () => {
    const serverSavedCodes = [];
    const v1Deferred = createDeferred();

    let editorVal = 'V0';
    let currentEx = { id: 1, initialCode: 'V0' };

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: 'V0',
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        getCurrentExercise: () => currentEx,
        apiSave: async (exId, code) => {
            if (code === 'V1') {
                await v1Deferred.promise;
            }
            serverSavedCodes.push(code);
            return { ok: true, status: 200 };
        }
    });

    // 1. Gõ V1 và kích hoạt lưu
    editorVal = 'V1';
    saveManager.queueDraftSave(1, 'V1', true);
    await new Promise(r => setTimeout(r, 5));

    // 2. Trong lúc V1 đang bay, người dùng sửa lại về V0 (giá trị đã lưu ban đầu)
    editorVal = 'V0';

    // 3. Gọi flush (hoặc điều hướng rời trang)
    const flushPromise = saveManager.flushCurrentDraft();

    // Cho V1 hoàn tất
    v1Deferred.resolve();
    await flushPromise;

    // Server phải kết thúc bằng V0 khớp với editor hiện tại
    assert.deepEqual(serverSavedCodes, ['V1', 'V0'], 'Server phải nhận lại V0 sau khi V1 kết thúc');
    assert.equal(saveManager.getLastSavedCode(1), 'V0', 'Bản cuối cùng xác nhận trên server là V0');
});

test('7. beforeunload khi flush chậm: Draft dirty, transition đang lưu -> Cảnh báo được kích hoạt, allowUnload chưa bật', async () => {
    let editorVal = 'dirty_code';
    let currentEx = { id: 1, initialCode: 'clean_code' };
    const flushDeferred = createDeferred();

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        getCurrentExercise: () => currentEx,
        apiSave: async () => {
            await flushDeferred.promise;
            return { ok: true, status: 200 };
        }
    });

    // Bắt đầu điều hướng an toàn (bắt đầu flush)
    const navPromise = saveManager.handleSafeNavigation('/dashboard', null);
    await new Promise(r => setTimeout(r, 5));

    // Lúc này request đang bay trên mạng
    assert.equal(saveManager.allowUnload, false, 'allowUnload PHẢI là false khi flush chưa xong');
    assert.equal(saveManager.hasUnconfirmedDrafts(), true, 'Còn draft chưa xác nhận');

    // Giả lập sự kiện beforeunload từ trình duyệt
    let defaultPrevented = false;
    let returnValue = undefined;
    const fakeEvent = {
        preventDefault: () => { defaultPrevented = true; },
        set returnValue(val) { returnValue = val; },
        get returnValue() { return returnValue; }
    };

    saveManager.handleBeforeUnload(fakeEvent);

    assert.equal(defaultPrevented, true, 'beforeunload phải gọi preventDefault() để chặn đóng tab');
    assert.equal(returnValue, '', 'returnValue phải được đặt chuỗi rỗng để kích hoạt popup trình duyệt');

    // Giải phóng flush
    flushDeferred.resolve();
    await navPromise;
});

test('8. beforeunload sau khi lưu đủ: Không còn pending/in-flight/dirty -> Không cảnh báo thừa', async () => {
    let editorVal = 'clean_code';
    let currentEx = { id: 1, initialCode: 'clean_code' };

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        getCurrentExercise: () => currentEx,
        apiSave: async () => ({ ok: true, status: 200 })
    });

    assert.equal(saveManager.hasUnconfirmedDrafts(), false, 'Không còn draft nào chưa xác nhận');

    let defaultPrevented = false;
    let returnValue = undefined;
    const fakeEvent = {
        preventDefault: () => { defaultPrevented = true; },
        set returnValue(val) { returnValue = val; },
        get returnValue() { return returnValue; }
    };

    const res = saveManager.handleBeforeUnload(fakeEvent);

    assert.equal(defaultPrevented, false, 'Không được gọi preventDefault khi mọi thứ đã lưu');
    assert.equal(returnValue, undefined, 'returnValue phải là undefined');
    assert.equal(res, undefined);
});

test('9. Lưu lỗi hoặc hủy rời trang: Flush lỗi, người dùng không bỏ draft -> Giữ code, mở khóa, guard vẫn hoạt động', async () => {
    let editorVal = 'valuable_code';
    let readOnlyState = false;
    let currentEx = { id: 1, initialCode: 'starter' };

    let userConfirmChoice = false; // Người dùng bấm Cancel (không chấp nhận mất code)

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx,
        confirmLeave: () => userConfirmChoice,
        apiSave: async () => {
            return { ok: false, status: 500 };
        }
    });

    let navigated = false;
    saveManager.onNavigate = () => { navigated = true; };

    const result = await saveManager.handleSafeNavigation('/home', null);

    assert.equal(result, false, 'Điều hướng bị hủy');
    assert.equal(navigated, false, 'Không thực hiện chuyển trang');
    assert.equal(saveManager.isTransitionInProgress(), false, 'Transition lock phải được mở lại');
    assert.equal(saveManager.allowUnload, false, 'allowUnload vẫn là false');
    assert.equal(readOnlyState, false, 'Editor được mở lại cho người dùng');
    assert.equal(editorVal, 'valuable_code', 'Code của người dùng được bảo toàn');
    assert.equal(saveManager.hasUnconfirmedDrafts(), true, 'Vẫn còn draft chưa lưu');

    // Thử lại lần 2 với xác nhận cho phép bỏ qua
    userConfirmChoice = true;
    const result2 = await saveManager.handleSafeNavigation('/home', null);
    assert.equal(result2, true, 'Cho phép rời trang sau khi người dùng đã xác nhận bỏ thay đổi');
    assert.equal(saveManager.allowUnload, true, 'allowUnload được bật');
});

test('10. Double-click / link / logout cạnh tranh: Chỉ một transition được thực hiện', async () => {
    let editorVal = 'some_code';
    let currentEx = { id: 1, initialCode: 'init' };
    const flushDeferred = createDeferred();
    let navCount = 0;

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        getCurrentExercise: () => currentEx,
        onNavigate: () => { navCount++; },
        apiSave: async () => {
            await flushDeferred.promise;
            return { ok: true, status: 200 };
        }
    });

    // Click lần 1 vào Back
    const nav1Promise = saveManager.handleSafeNavigation('/prev', null);

    // Click lần 2 ngay sau đó vào Home hoặc Logout trong lúc lần 1 đang flush
    const nav2Promise = saveManager.handleSafeNavigation('/home', null);

    assert.equal(await nav2Promise, false, 'Thao tác điều hướng thứ hai phải bị từ chối ngay lập tức');

    flushDeferred.resolve();
    assert.equal(await nav1Promise, true, 'Thao tác điều hướng đầu tiên hoàn thành');
    assert.equal(navCount, 1, 'Chỉ thực hiện điều hướng đúng 1 lần');
});

test('11. Khởi tạo exercise đầu tiên khi request draft đang chạy: Editor và thao tác đổi code bị khóa', async () => {
    let editorVal = '';
    let readOnlyState = false;
    let currentEx = null;

    const saveManager = createSaveManager({
        initialExercise: null,
        initialCode: '',
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx
    });

    const draftDeferred = createDeferred();

    const ex1 = { id: 101, initialCode: 'print("ex101_init")' };
    const initPromise = saveManager.initExercise(ex1, {
        render: (ex) => {
            currentEx = ex;
            editorVal = ex.initialCode;
        },
        fetchDraft: async () => draftDeferred.promise
    });

    await new Promise(r => setTimeout(r, 5));

    // Trong khi draft request đang bay:
    assert.equal(saveManager.isTransitionInProgress(), true, 'Transition lock phải bật khi đang khởi tạo');
    assert.equal(readOnlyState, true, 'Editor phải ở chế độ readOnly');

    // Chặn reset
    function tryReset() {
        if (saveManager.isTransitionInProgress()) return false;
        editorVal = currentEx.initialCode;
        return true;
    }
    assert.equal(tryReset(), false, 'Reset code phải bị chặn');

    // Chặn transition cạnh tranh
    assert.equal(saveManager.startTransition(), false, 'Không được phép bắt đầu transition cạnh tranh');

    draftDeferred.resolve({ ok: true, code: 'print("ex101_draft")' });
    await initPromise;
});

test('12. Draft trả về thành công: Nạp đúng nội dung rồi mới mở khóa', async () => {
    let editorVal = '';
    let readOnlyState = false;
    let currentEx = null;

    const saveManager = createSaveManager({
        initialExercise: null,
        initialCode: '',
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx
    });

    const ex = { id: 202, initialCode: 'print("init_202")' };
    const res = await saveManager.initExercise(ex, {
        render: (e) => {
            currentEx = e;
            editorVal = e.initialCode;
        },
        fetchDraft: async () => ({ ok: true, code: 'print("loaded_draft_202")' })
    });

    assert.equal(res.ok, true);
    assert.equal(editorVal, 'print("loaded_draft_202")', 'Mã nguồn editor phải là bản nháp đã nạp');
    assert.equal(saveManager.getLastSavedCode(202), 'print("loaded_draft_202")', 'Mã đã lưu phải khớp draft');
    assert.equal(readOnlyState, false, 'Editor phải được mở khóa sau khi nạp xong');
    assert.equal(saveManager.isTransitionInProgress(), false, 'Transition lock đã được giải phóng');
    assert.equal(saveManager.saveStatus, 'saved', 'Trạng thái lưu là saved');
});

test('13. Request draft lỗi: Xử lý đúng trạng thái error, mở khóa an toàn và cho phép retry', async () => {
    let editorVal = '';
    let readOnlyState = false;
    let currentEx = null;

    const saveManager = createSaveManager({
        initialExercise: null,
        initialCode: '',
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx
    });

    const ex = { id: 303, initialCode: 'print("init_303")' };
    const res = await saveManager.initExercise(ex, {
        render: (e) => {
            currentEx = e;
            editorVal = e.initialCode;
        },
        fetchDraft: async () => ({ ok: false, error: new Error('HTTP 500 Internal Server Error') })
    });

    assert.equal(res.ok, false, 'initExercise phải báo lỗi');
    assert.equal(saveManager.saveStatus, 'error', 'Trạng thái UI phải là error (không được báo saved)');
    assert.equal(readOnlyState, false, 'Editor không bị khóa vĩnh viễn');
    assert.equal(saveManager.isTransitionInProgress(), false, 'Transition lock được giải phóng sau lỗi');

    // Thử lại (retry) thành công
    const retryRes = await saveManager.retryLoadDraft(303, async () => ({
        ok: true,
        code: 'print("recovered_draft_303")'
    }));

    assert.equal(retryRes.ok, true, 'Retry nạp draft thành công');
    assert.equal(editorVal, 'print("recovered_draft_303")', 'Mã nguồn editor được khôi phục');
    assert.equal(saveManager.saveStatus, 'saved', 'Trạng thái chuyển sang saved');
    assert.equal(readOnlyState, false, 'Editor mở khóa sau retry');
});

test('14. Response cũ (stale response): Không ghi đè exercise hiện tại', async () => {
    let editorVal = '';
    let readOnlyState = false;
    let currentEx = null;

    const saveManager = createSaveManager({
        initialExercise: null,
        initialCode: '',
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx
    });

    const draftDeferred1 = createDeferred();
    const ex1 = { id: 1, initialCode: 'init_1' };

    // Khởi tạo bài 1 (request draft bị treo)
    const init1Promise = saveManager.initExercise(ex1, {
        render: (e) => { currentEx = e; editorVal = e.initialCode; },
        fetchDraft: async () => draftDeferred1.promise
    });

    await new Promise(r => setTimeout(r, 5));

    // Người dùng chuyển sang bài 2
    saveManager.endTransition(); // giải phóng để bài 2 chạy
    const ex2 = { id: 2, initialCode: 'init_2' };
    await saveManager.initExercise(ex2, {
        render: (e) => { currentEx = e; editorVal = e.initialCode; },
        fetchDraft: async () => ({ ok: true, code: 'draft_2_active' })
    });

    assert.equal(editorVal, 'draft_2_active', 'Editor đang hiển thị draft bài 2');

    // Lúc này response của bài 1 mới về (stale response)
    draftDeferred1.resolve({ ok: true, code: 'stale_draft_1' });
    const res1 = await init1Promise;

    assert.equal(res1.discarded, true, 'Response cũ phải bị đánh dấu discarded');
    assert.equal(editorVal, 'draft_2_active', 'Editor KHÔNG bị response cũ của bài 1 ghi đè');
    assert.equal(currentEx.id, 2, 'Context vẫn là bài 2');
});

test('15. Luồng loadExercise() giữ khóa xuyên suốt cho đến khi nạp xong draft', async () => {
    let editorVal = 'code_A';
    let readOnlyState = false;
    let currentEx = { id: 1, initialCode: 'code_A' };

    const saveManager = createSaveManager({
        initialExercise: currentEx,
        initialCode: currentEx.initialCode,
        getEditorValue: () => editorVal,
        setEditorValue: (v) => { editorVal = v; },
        setReadOnly: (ro) => { readOnlyState = ro; },
        isReadOnly: () => readOnlyState,
        getCurrentExercise: () => currentEx,
        apiSave: async () => ({ ok: true, status: 200 })
    });

    const draftDeferredB = createDeferred();

    async function loadExerciseWithDraft(targetEx) {
        if (!saveManager.startTransition()) return false;
        const gen = saveManager.getNextGeneration();
        try {
            // Render bài mới
            currentEx = targetEx;
            editorVal = targetEx.initialCode;

            // Chờ draft nạp xong TRƯỚC KHI mở khóa
            const draftRes = await draftDeferredB.promise;
            if (gen === saveManager.currentGeneration && draftRes.ok) {
                editorVal = draftRes.code;
                saveManager.setLastSavedCode(targetEx.id, draftRes.code);
            }
            return true;
        } finally {
            saveManager.endTransition();
        }
    }

    const loadPromise = loadExerciseWithDraft({ id: 2, initialCode: 'init_B' });
    await new Promise(r => setTimeout(r, 5));

    // Trong khi draft bài B đang tải
    assert.equal(saveManager.isTransitionInProgress(), true, 'Khóa transition phải bật');
    assert.equal(readOnlyState, true, 'Editor phải khóa trong suốt quá trình');

    draftDeferredB.resolve({ ok: true, code: 'draft_B_final' });
    await loadPromise;

    assert.equal(saveManager.isTransitionInProgress(), false, 'Khóa mở sau khi hoàn tất');
    assert.equal(readOnlyState, false, 'Editor mở lại');
    assert.equal(editorVal, 'draft_B_final', 'Mã nguồn là draft B');
});
