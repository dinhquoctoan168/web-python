import test from 'node:test';
import assert from 'node:assert/strict';

/**
 * Mô phỏng Save Worker tuần tự theo đúng thiết kế nâng cấp của web/static/js/ide.js
 */
function createSaveWorker(apiHandler) {
    const lastSavedCodeByExercise = {};
    let isWorkerRunning = false;
    let pendingSnapshot = null;
    let inFlightSnapshot = null;
    let autoSaveTimer = null;
    let lastSaveFailed = false;
    let workerWaiters = [];
    let saveStatus = 'saved';
    let inFlightRequests = 0;
    let maxConcurrentRequests = 0;

    let currentExercise = { id: 1, initialCode: 'print("hello")' };
    let editorValue = currentExercise.initialCode;
    lastSavedCodeByExercise[1] = currentExercise.initialCode;

    function hasUnsavedChanges() {
        const savedCode = lastSavedCodeByExercise[currentExercise.id];
        return editorValue !== savedCode;
    }

    function updateSaveStatusUI() {
        if (lastSaveFailed) {
            saveStatus = 'error';
        } else if (isWorkerRunning || pendingSnapshot || autoSaveTimer) {
            saveStatus = 'saving';
        } else if (hasUnsavedChanges()) {
            saveStatus = 'saving';
        } else {
            saveStatus = 'saved';
        }
    }

    function queueDraftSave(exerciseId, code, immediate = false) {
        lastSaveFailed = false;
        if (inFlightSnapshot && inFlightSnapshot.exerciseId === exerciseId && inFlightSnapshot.code === code && !pendingSnapshot) {
            // Snapshot giống hệt đang bay, không cần tạo pending trùng lặp
        } else {
            pendingSnapshot = { exerciseId, code };
        }
        updateSaveStatusUI();

        if (autoSaveTimer) {
            clearTimeout(autoSaveTimer);
            autoSaveTimer = null;
        }

        if (immediate) {
            kickSaveWorker();
        } else {
            autoSaveTimer = setTimeout(() => {
                autoSaveTimer = null;
                kickSaveWorker();
            }, 30);
        }
    }

    async function kickSaveWorker() {
        if (isWorkerRunning) return;
        isWorkerRunning = true;
        updateSaveStatusUI();

        while (pendingSnapshot) {
            const currentItem = pendingSnapshot;
            pendingSnapshot = null;
            inFlightSnapshot = currentItem;

            inFlightRequests++;
            if (inFlightRequests > maxConcurrentRequests) {
                maxConcurrentRequests = inFlightRequests;
            }

            try {
                const res = await apiHandler(currentItem.exerciseId, currentItem.code);
                inFlightRequests--;

                if (!res.ok) throw new Error('HTTP ' + res.status);
                lastSavedCodeByExercise[currentItem.exerciseId] = currentItem.code;
                lastSaveFailed = false;
            } catch (err) {
                inFlightRequests--;
                lastSaveFailed = true;
                if (!pendingSnapshot) {
                    pendingSnapshot = currentItem;
                }
                break;
            } finally {
                inFlightSnapshot = null;
            }
        }

        isWorkerRunning = false;
        updateSaveStatusUI();

        const waiters = workerWaiters;
        workerWaiters = [];
        waiters.forEach(resolve => resolve(!hasUnsavedChanges()));
    }

    async function flushCurrentDraft() {
        if (autoSaveTimer) {
            clearTimeout(autoSaveTimer);
            autoSaveTimer = null;
        }

        if (hasUnsavedChanges()) {
            if (!inFlightSnapshot || inFlightSnapshot.code !== editorValue || inFlightSnapshot.exerciseId !== currentExercise.id) {
                pendingSnapshot = { exerciseId: currentExercise.id, code: editorValue };
            }
        }

        if (!pendingSnapshot && !isWorkerRunning) {
            return !hasUnsavedChanges();
        }

        const waitPromise = new Promise(resolve => {
            workerWaiters.push(resolve);
        });

        kickSaveWorker();
        return await waitPromise;
    }

    return {
        get editorValue() { return editorValue; },
        set editorValue(v) { editorValue = v; },
        get saveStatus() { return saveStatus; },
        get maxConcurrentRequests() { return maxConcurrentRequests; },
        get lastSavedCode() { return lastSavedCodeByExercise[currentExercise.id]; },
        hasUnsavedChanges,
        queueDraftSave,
        flushCurrentDraft,
        setExercise: (ex) => {
            currentExercise = ex;
            editorValue = ex.initialCode;
            if (!lastSavedCodeByExercise[ex.id]) {
                lastSavedCodeByExercise[ex.id] = ex.initialCode;
            }
        }
    };
}

test('1. Ghi tuần tự: Không có 2 request lưu chạy song song và bản cuối cùng là V3', async () => {
    const serverLogs = [];
    let resolveV1;
    const v1Promise = new Promise(res => { resolveV1 = res; });

    const worker = createSaveWorker(async (exerciseId, code) => {
        serverLogs.push({ event: 'start', code });
        if (code === 'V1') {
            await v1Promise;
        } else {
            await new Promise(r => setTimeout(r, 10));
        }
        serverLogs.push({ event: 'finish', code });
        return { ok: true, status: 200 };
    });

    // 1. Gõ V1 và kích hoạt lưu ngay
    worker.editorValue = 'V1';
    worker.queueDraftSave(1, 'V1', true);

    // Đợi chút để V1 bắt đầu gửi
    await new Promise(r => setTimeout(r, 5));

    // 2. Trong lúc V1 đang bay, người dùng gõ V2 rồi V3
    worker.editorValue = 'V2';
    worker.queueDraftSave(1, 'V2', true);

    worker.editorValue = 'V3';
    worker.queueDraftSave(1, 'V3', true);

    // Cho V1 hoàn thành
    resolveV1();

    // Chờ flush hoàn tất
    const success = await worker.flushCurrentDraft();

    assert.equal(success, true, 'Flush phải thành công');
    assert.equal(worker.maxConcurrentRequests, 1, 'Số request đồng thời tối đa phải là 1 (Single-flight)');
    assert.equal(worker.lastSavedCode, 'V3', 'Bản ghi cuối cùng được xác nhận phải là V3');
    assert.equal(worker.hasUnsavedChanges(), false, 'Không còn thay đổi chưa lưu');

    // Kiểm tra thứ tự server nhận
    const finishedCodes = serverLogs.filter(l => l.event === 'finish').map(l => l.code);
    assert.deepEqual(finishedCodes, ['V1', 'V3'], 'Server chỉ nhận V1 rồi đến bản mới nhất V3 (V2 đã được gộp)');
});

test('2. Nhiều lời gọi flushCurrentDraft đồng thời không tạo request song song', async () => {
    let callCount = 0;
    const worker = createSaveWorker(async (exerciseId, code) => {
        callCount++;
        await new Promise(r => setTimeout(r, 20));
        return { ok: true, status: 200 };
    });

    worker.editorValue = 'code_ABC';
    worker.queueDraftSave(1, 'code_ABC', false);

    // Gọi flush đồng thời từ 3 nơi (Back button, AutoSave, Menu)
    const [res1, res2, res3] = await Promise.all([
        worker.flushCurrentDraft(),
        worker.flushCurrentDraft(),
        worker.flushCurrentDraft()
    ]);

    assert.equal(res1, true);
    assert.equal(res2, true);
    assert.equal(res3, true);
    assert.equal(worker.maxConcurrentRequests, 1, 'Chỉ 1 request chạy tại 1 thời điểm');
    assert.equal(callCount, 1, 'Chỉ gọi API 1 lần cho snapshot đó');
});

test('3. API lỗi thì giữ nguyên code, flush trả false, sau đó retry thành công', async () => {
    let shouldFail = true;
    const worker = createSaveWorker(async (exerciseId, code) => {
        await new Promise(r => setTimeout(r, 10));
        if (shouldFail) {
            return { ok: false, status: 500 };
        }
        return { ok: true, status: 200 };
    });

    worker.editorValue = 'draft_important';
    worker.queueDraftSave(1, 'draft_important', true);

    const firstFlush = await worker.flushCurrentDraft();
    assert.equal(firstFlush, false, 'Lưu thất bại phải trả false');
    assert.equal(worker.saveStatus, 'error', 'Trạng thái UI phải là error');
    assert.equal(worker.hasUnsavedChanges(), true, 'Code vẫn chưa được đánh dấu là saved');

    // Mạng phục hồi
    shouldFail = false;
    const retryFlush = await worker.flushCurrentDraft();
    assert.equal(retryFlush, true, 'Retry thành công phải trả true');
    assert.equal(worker.saveStatus, 'saved', 'Trạng thái UI trở lại saved');
    assert.equal(worker.lastSavedCode, 'draft_important');
});

test('4. Chặn chuyển bài khi flush lưu thất bại', async () => {
    const worker = createSaveWorker(async () => {
        return { ok: false, status: 500 };
    });

    worker.editorValue = 'my_exercise_1_code';

    // Giả lập logic chuyển bài của window.loadExercise
    let currentExId = 1;
    let switched = false;

    async function trySwitchExercise(newId) {
        if (worker.hasUnsavedChanges()) {
            const saved = await worker.flushCurrentDraft();
            if (!saved || worker.hasUnsavedChanges()) {
                return false; // Dừng chuyển bài
            }
        }
        currentExId = newId;
        switched = true;
        return true;
    }

    const switchResult = await trySwitchExercise(2);
    assert.equal(switchResult, false, 'Chuyển bài phải bị từ chối');
    assert.equal(switched, false, 'Không được phép chuyển bài');
    assert.equal(currentExId, 1, 'Vẫn giữ bài tập cũ');
    assert.equal(worker.editorValue, 'my_exercise_1_code', 'Code bài cũ còn nguyên');
});
