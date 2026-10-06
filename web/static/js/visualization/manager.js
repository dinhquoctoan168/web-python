/**
 * Visualization Manager (Vanilla JS, Zero Framework)
 * Điều khiển và đồng bộ giao diện mô phỏng thuật toán
 */

window.VisualizationManager = (function() {
    let currentVisualizer = null;
    let autoPlayTimer = null;
    let autoPlayIntervalMs = 1200;

    const dom = {
        section: null,
        title: null,
        badge: null,
        stage: null,
        explanation: null,
        stepCounter: null,
        btnPrev: null,
        btnNext: null,
        btnReset: null,
        btnAutoPlay: null
    };

    function initDom() {
        dom.section = document.getElementById('visualizationSection');
        dom.title = document.getElementById('vizTitle');
        dom.badge = document.getElementById('vizBadge');
        dom.stage = document.getElementById('vizStage');
        dom.explanation = document.getElementById('vizExplanation');
        dom.stepCounter = document.getElementById('vizStepCounter');
        dom.btnPrev = document.getElementById('btnVizPrev');
        dom.btnNext = document.getElementById('btnVizNext');
        dom.btnReset = document.getElementById('btnVizReset');
        dom.btnAutoPlay = document.getElementById('btnVizAutoPlay');

        if (dom.btnPrev) dom.btnPrev.onclick = handlePrev;
        if (dom.btnNext) dom.btnNext.onclick = handleNext;
        if (dom.btnReset) dom.btnReset.onclick = handleReset;
        if (dom.btnAutoPlay) dom.btnAutoPlay.onclick = toggleAutoPlay;
    }

    function stopAutoPlay() {
        if (autoPlayTimer) {
            clearInterval(autoPlayTimer);
            autoPlayTimer = null;
        }
        if (dom.btnAutoPlay) {
            dom.btnAutoPlay.innerHTML = '<span class="icon">&#9658;</span> Tự động';
            dom.btnAutoPlay.classList.remove('viz-btn-primary');
            dom.btnAutoPlay.classList.add('viz-btn-secondary');
        }
    }

    function toggleAutoPlay() {
        if (autoPlayTimer) {
            stopAutoPlay();
        } else {
            if (!currentVisualizer) return;
            if (dom.btnAutoPlay) {
                dom.btnAutoPlay.innerHTML = '<span class="icon">&#9646;&#9646;</span> Tạm dừng';
                dom.btnAutoPlay.classList.remove('viz-btn-secondary');
                dom.btnAutoPlay.classList.add('viz-btn-primary');
            }
            autoPlayTimer = setInterval(function() {
                if (!currentVisualizer) {
                    stopAutoPlay();
                    return;
                }
                const cur = currentVisualizer.getCurrentIndex();
                const total = currentVisualizer.getTotalSteps();
                if (cur >= total - 1) {
                    stopAutoPlay();
                    return;
                }
                handleNext();
            }, autoPlayIntervalMs);
        }
    }

    function updateView() {
        if (!currentVisualizer) return;
        const state = currentVisualizer.getCurrentState();
        if (!state) return;

        currentVisualizer.render(dom.stage, state);

        if (dom.explanation && state.message) {
            dom.explanation.innerHTML = state.message;
        }

        const cur = currentVisualizer.getCurrentIndex();
        const total = currentVisualizer.getTotalSteps();

        if (dom.stepCounter) {
            dom.stepCounter.textContent = `${cur + 1} / ${total}`;
        }

        if (dom.btnPrev) dom.btnPrev.disabled = (cur === 0);
        if (dom.btnNext) dom.btnNext.disabled = (cur >= total - 1);
    }

    function handleNext() {
        if (!currentVisualizer) return;
        currentVisualizer.next();
        updateView();
    }

    function handlePrev() {
        stopAutoPlay();
        if (!currentVisualizer) return;
        currentVisualizer.prev();
        updateView();
    }

    function handleReset() {
        stopAutoPlay();
        if (!currentVisualizer) return;
        currentVisualizer.reset();
        updateView();
    }

    return {
        load: function(exercise) {
            if (!dom.section) initDom();
            stopAutoPlay();

            const vizType = (exercise && (exercise.visualization_type || exercise.visualizationType)) ? (exercise.visualization_type || exercise.visualizationType).toLowerCase() : '';

            if (!vizType) {
                this.hide();
                return;
            }

            if (vizType === 'binary_search' && window.BinarySearchVisualizer) {
                currentVisualizer = window.BinarySearchVisualizer;
                if (dom.title) dom.title.textContent = 'Mô phỏng Tìm kiếm Nhị phân (Binary Search)';
                currentVisualizer.init();
            } else if (vizType === 'array' && window.ArrayVisualizer) {
                currentVisualizer = window.ArrayVisualizer;
                if (dom.title) dom.title.textContent = 'Mô phỏng Thao tác trên Mảng (Array)';
                currentVisualizer.init();
            } else if (vizType === 'stack' && window.StackVisualizer) {
                currentVisualizer = window.StackVisualizer;
                if (dom.title) dom.title.textContent = 'Mô phỏng Ngăn xếp (Stack - LIFO)';
                currentVisualizer.init();
            } else if ((vizType === 'sorting' || vizType === 'sort') && window.SortingVisualizer) {
                currentVisualizer = window.SortingVisualizer;
                if (dom.title) dom.title.textContent = 'Mô phỏng Thuật toán Sắp xếp (Sorting)';
                currentVisualizer.init();
            } else {
                this.hide();
                return;
            }

            if (dom.section) dom.section.style.display = 'block';
            updateView();
        },
        hide: function() {
            stopAutoPlay();
            currentVisualizer = null;
            if (dom.section) dom.section.style.display = 'none';
        }
    };
})();
