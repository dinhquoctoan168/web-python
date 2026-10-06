/**
 * Sorting Visualizer (Vanilla JS, Zero Framework)
 * Mô phỏng Thuật toán Sắp xếp (Bubble Sort) từng bước
 */

window.SortingVisualizer = (function() {
    let steps = [];
    let currentStepIndex = 0;

    function buildSteps(initialArr) {
        const result = [];
        const arr = [...initialArr];
        const n = arr.length;
        let stepCount = 0;

        result.push({
            array: [...arr],
            comparing: [-1, -1],
            sortedIndices: [],
            step: 0,
            message: `Khởi tạo: Sắp xếp mảng [${arr.join(', ')}] theo thứ tự tăng dần.`
        });

        const sorted = [];

        for (let i = 0; i < n - 1; i++) {
            for (let j = 0; j < n - i - 1; j++) {
                stepCount++;
                const compA = arr[j];
                const compB = arr[j + 1];

                if (compA > compB) {
                    // Cần đổi chỗ
                    arr[j] = compB;
                    arr[j + 1] = compA;

                    result.push({
                        array: [...arr],
                        comparing: [j, j + 1],
                        sortedIndices: [...sorted],
                        step: stepCount,
                        message: `Bước ${stepCount}: So sánh <strong>arr[${j}] (${compA})</strong> &gt; <strong>arr[${j + 1}] (${compB})</strong> &rarr; Đổi chỗ hai phần tử.`
                    });
                } else {
                    result.push({
                        array: [...arr],
                        comparing: [j, j + 1],
                        sortedIndices: [...sorted],
                        step: stepCount,
                        message: `Bước ${stepCount}: So sánh <strong>arr[${j}] (${compA})</strong> &le; <strong>arr[${j + 1}] (${compB})</strong> &rarr; Giữ nguyên vị trí.`
                    });
                }
            }
            sorted.push(n - i - 1);
        }
        sorted.push(0);

        result.push({
            array: [...arr],
            comparing: [-1, -1],
            sortedIndices: Array.from({ length: n }, (_, i) => i),
            step: stepCount + 1,
            message: `Hoàn tất: Mảng đã được sắp xếp tăng dần hoàn chỉnh: [${arr.join(', ')}].`
        });

        return result;
    }

    function render(container, state) {
        if (!container) return;

        const maxVal = Math.max(...state.array, 1);
        let html = `
            <div class="viz-sort-bars">
        `;

        state.array.forEach((val, idx) => {
            const heightPercent = Math.max(15, Math.round((val / maxVal) * 100));
            const isComparing = (state.comparing[0] === idx || state.comparing[1] === idx);
            const isSorted = state.sortedIndices.includes(idx);

            let barClass = 'viz-bar';
            if (isComparing) barClass += ' comparing';
            if (isSorted) barClass += ' sorted';

            html += `
                <div class="viz-bar-wrapper">
                    <div class="${barClass}" style="height: ${heightPercent}px;"></div>
                    <span class="viz-bar-val">${val}</span>
                </div>
            `;
        });

        html += `
            </div>
        `;

        container.innerHTML = html;
    }

    return {
        init: function(customArray) {
            const arr = customArray && customArray.length > 0 ? customArray : [45, 18, 62, 29, 83, 11, 37];
            steps = buildSteps(arr);
            currentStepIndex = 0;
            return this.getCurrentState();
        },
        getCurrentState: function() {
            return steps[currentStepIndex] || null;
        },
        getCurrentIndex: function() {
            return currentStepIndex;
        },
        getTotalSteps: function() {
            return steps.length;
        },
        next: function() {
            if (currentStepIndex < steps.length - 1) {
                currentStepIndex++;
            }
            return this.getCurrentState();
        },
        prev: function() {
            if (currentStepIndex > 0) {
                currentStepIndex--;
            }
            return this.getCurrentState();
        },
        reset: function() {
            currentStepIndex = 0;
            return this.getCurrentState();
        },
        render: render
    };
})();
