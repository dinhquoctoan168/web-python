/**
 * IdeSaveWorker - Save Worker tuần tự & Transition Guard Module
 * Hỗ trợ UMD: nạp trực tiếp qua browser script hoặc import trong Node.js test runner.
 */
(function (root, factory) {
    if (typeof exports === 'object' && typeof module !== 'undefined') {
        module.exports = factory();
    } else if (typeof define === 'function' && define.amd) {
        define(factory);
    } else {
        root.IdeSaveWorker = factory();
    }
}(typeof globalThis !== 'undefined' ? globalThis : this, function () {
    'use strict';

    function createSaveManager(options = {}) {
        const apiSave = options.apiSave || (async () => ({ ok: true, status: 200 }));
        let editorValue = options.initialCode || '';
        let currentExercise = options.initialExercise || null;
        let isReadOnly = false;

        const getEditorValueFn = options.getEditorValue || (() => editorValue);
        const setEditorValueFn = options.setEditorValue || ((v) => { editorValue = v; });
        const setReadOnlyFn = options.setReadOnly || ((ro) => { isReadOnly = ro; });
        const isReadOnlyFn = options.isReadOnly || (() => isReadOnly);
        const getCurrentExerciseFn = options.getCurrentExercise || (() => currentExercise);

        const onStatusChange = options.onStatusChange || (() => {});
        const onExerciseStatusUpdate = options.onExerciseStatusUpdate || (() => {});
        const onNavigate = options.onNavigate || ((targetUrl, form) => {});
        const confirmLeave = options.confirmLeave || (() => true);

        const debounceMs = typeof options.debounceMs === 'number' ? options.debounceMs : 1500;
        const timerHost = options.timers || globalThis;

        const lastSavedCodeByExercise = {};
        if (currentExercise) {
            lastSavedCodeByExercise[currentExercise.id] = currentExercise.initialCode || '';
        }

        let isWorkerRunning = false;
        let inFlightSnapshot = null; // { exerciseId, code }
        const pendingMap = new Map(); // exerciseId -> code
        let autoSaveTimer = null;
        let lastSaveFailed = false;
        let workerWaiters = [];
        let inFlightRequests = 0;
        let maxConcurrentRequests = 0;
        let saveStatus = 'saved';

        let transitionInProgress = false;
        let allowUnload = false;

        function getLastSavedCode(exerciseId) {
            const exId = exerciseId !== undefined ? exerciseId : (currentExercise ? currentExercise.id : null);
            if (!exId) return undefined;
            return lastSavedCodeByExercise[exId];
        }

        function setLastSavedCode(exerciseId, code) {
            lastSavedCodeByExercise[exerciseId] = code;
            updateSaveStatusUI();
        }

        function hasUnsavedChanges() {
            const ex = getCurrentExerciseFn();
            if (!ex) return false;
            const currentCode = getEditorValueFn() || '';
            const savedCode = lastSavedCodeByExercise[ex.id];
            if (savedCode === undefined) {
                return currentCode !== (ex.initialCode || '');
            }
            return currentCode !== savedCode;
        }

        function hasUnconfirmedDrafts() {
            const ex = getCurrentExerciseFn();
            const currentCode = getEditorValueFn() || '';

            // 1. Khung soạn thảo của bài hiện tại có thay đổi chưa lưu
            if (hasUnsavedChanges()) return true;

            // 2. Timer debounce đang đếm lùi
            if (autoSaveTimer !== null) return true;

            // 3. Còn bản nháp trong hàng đợi pending (của bất kỳ bài nào)
            if (pendingMap.size > 0) return true;

            // 4. Đang có request lưu ngầm trên mạng
            if (inFlightSnapshot !== null || isWorkerRunning) return true;

            return false;
        }

        function updateSaveStatusUI() {
            const ex = getCurrentExerciseFn();
            if (!ex) return;

            if (lastSaveFailed) {
                saveStatus = 'error';
            } else if (isWorkerRunning || pendingMap.size > 0 || autoSaveTimer !== null) {
                saveStatus = 'saving';
            } else if (hasUnsavedChanges()) {
                saveStatus = 'saving';
            } else {
                saveStatus = 'saved';
            }
            onStatusChange(saveStatus);
        }

        function queueDraftSave(exerciseId, code, immediate = false) {
            lastSaveFailed = false;

            if (inFlightSnapshot && inFlightSnapshot.exerciseId === exerciseId && inFlightSnapshot.code === code && !pendingMap.has(exerciseId)) {
                // Snapshot giống hệt đang được gửi, không tạo pending trùng lặp
            } else {
                pendingMap.set(exerciseId, code);
            }
            updateSaveStatusUI();

            if (autoSaveTimer !== null) {
                timerHost.clearTimeout(autoSaveTimer);
                autoSaveTimer = null;
            }

            if (immediate) {
                kickSaveWorker();
            } else {
                autoSaveTimer = timerHost.setTimeout(() => {
                    autoSaveTimer = null;
                    kickSaveWorker();
                }, debounceMs);
            }
        }

        async function kickSaveWorker() {
            if (isWorkerRunning) return;
            isWorkerRunning = true;
            updateSaveStatusUI();

            while (pendingMap.size > 0) {
                const [nextExerciseId, nextCode] = pendingMap.entries().next().value;
                pendingMap.delete(nextExerciseId);

                const currentItem = { exerciseId: nextExerciseId, code: nextCode };
                inFlightSnapshot = currentItem;

                inFlightRequests++;
                if (inFlightRequests > maxConcurrentRequests) {
                    maxConcurrentRequests = inFlightRequests;
                }

                try {
                    const res = await apiSave(currentItem.exerciseId, currentItem.code);
                    inFlightRequests--;

                    if (!res || !res.ok) {
                        throw new Error('HTTP ' + (res ? res.status : 500));
                    }

                    lastSavedCodeByExercise[currentItem.exerciseId] = currentItem.code;
                    lastSaveFailed = false;

                    if (res.data && res.data.status) {
                        onExerciseStatusUpdate(currentItem.exerciseId, res.data.status);
                    }
                } catch (err) {
                    inFlightRequests--;
                    lastSaveFailed = true;

                    // Nếu lưu thất bại và chưa có snapshot mới hơn được đẩy vào, re-queue để retry
                    if (!pendingMap.has(currentItem.exerciseId)) {
                        pendingMap.set(currentItem.exerciseId, currentItem.code);
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
            const isClean = !hasUnconfirmedDrafts();
            waiters.forEach(resolve => resolve(isClean));
        }

        async function flushCurrentDraft() {
            if (autoSaveTimer !== null) {
                timerHost.clearTimeout(autoSaveTimer);
                autoSaveTimer = null;
            }

            const ex = getCurrentExerciseFn();
            if (ex) {
                const currentCode = getEditorValueFn() || '';
                const savedCode = lastSavedCodeByExercise[ex.id] !== undefined
                    ? lastSavedCodeByExercise[ex.id]
                    : (ex.initialCode || '');

                const inFlightDiffers = inFlightSnapshot && inFlightSnapshot.exerciseId === ex.id && inFlightSnapshot.code !== currentCode;
                const codeDiffers = currentCode !== savedCode;

                if (codeDiffers || inFlightDiffers) {
                    pendingMap.set(ex.id, currentCode);
                }
            }

            if (pendingMap.size === 0 && !isWorkerRunning) {
                return !hasUnconfirmedDrafts();
            }

            const waitPromise = new Promise(resolve => {
                workerWaiters.push(resolve);
            });

            kickSaveWorker();
            return await waitPromise;
        }

        function startTransition() {
            if (transitionInProgress) return false;
            transitionInProgress = true;
            setReadOnlyFn(true);
            updateSaveStatusUI();
            return true;
        }

        function endTransition() {
            transitionInProgress = false;
            setReadOnlyFn(false);
            updateSaveStatusUI();
        }

        async function handleSafeNavigation(targetUrl, formToSubmit) {
            if (transitionInProgress) return false;
            transitionInProgress = true;
            setReadOnlyFn(true);

            try {
                if (hasUnconfirmedDrafts()) {
                    updateSaveStatusUI();
                    const success = await flushCurrentDraft();
                    if (!success || hasUnconfirmedDrafts()) {
                        // Lưu lỗi
                        setReadOnlyFn(false);
                        transitionInProgress = false;
                        updateSaveStatusUI();

                        const proceed = confirmLeave('Không thể lưu mã nguồn mới nhất lên máy chủ (mất mạng hoặc lỗi kết nối). Bạn có chắc chắn muốn rời trang và chấp nhận mất phần thay đổi chưa lưu?');
                        if (proceed) {
                            allowUnload = true;
                            onNavigate(targetUrl, formToSubmit);
                            return true;
                        }
                        return false;
                    }
                }

                allowUnload = true;
                onNavigate(targetUrl, formToSubmit);
                return true;
            } catch (err) {
                setReadOnlyFn(false);
                transitionInProgress = false;
                allowUnload = false;
                updateSaveStatusUI();
                return false;
            }
        }

        function handleBeforeUnload(event) {
            if (!allowUnload && hasUnconfirmedDrafts()) {
                if (event) {
                    if (typeof event.preventDefault === 'function') {
                        event.preventDefault();
                    }
                    event.returnValue = '';
                }
                return '';
            }
            return undefined;
        }

        function setExercise(ex) {
            currentExercise = ex;
            editorValue = ex ? (ex.initialCode || '') : '';
            setEditorValueFn(editorValue);
            if (ex && lastSavedCodeByExercise[ex.id] === undefined) {
                lastSavedCodeByExercise[ex.id] = ex.initialCode || '';
            }
            updateSaveStatusUI();
        }

        return {
            get editorValue() { return getEditorValueFn(); },
            set editorValue(v) { setEditorValueFn(v); editorValue = v; updateSaveStatusUI(); },
            get saveStatus() { return saveStatus; },
            get maxConcurrentRequests() { return maxConcurrentRequests; },
            get inFlightRequests() { return inFlightRequests; },
            get inFlightSnapshot() { return inFlightSnapshot; },
            get isWorkerRunning() { return isWorkerRunning; },
            get pendingCount() { return pendingMap.size; },
            get lastSavedCode() {
                const ex = getCurrentExerciseFn();
                return ex ? lastSavedCodeByExercise[ex.id] : undefined;
            },
            get transitionInProgress() { return transitionInProgress; },
            get allowUnload() { return allowUnload; },
            set allowUnload(v) { allowUnload = v; },

            getLastSavedCode,
            setLastSavedCode,
            hasUnsavedChanges,
            hasUnconfirmedDrafts,
            queueDraftSave,
            flushCurrentDraft,
            startTransition,
            endTransition,
            handleSafeNavigation,
            handleBeforeUnload,
            setExercise,
            updateSaveStatusUI,
            isTransitionInProgress: () => transitionInProgress
        };
    }

    return {
        createSaveManager
    };
}));
