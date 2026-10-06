/**
 * Array Operations Visualizer (Vanilla JS, Zero Framework)
 * Mô phỏng Duyệt mảng & Tìm kiếm tuyến tính (Linear Search)
 */

window.ArrayVisualizer = (function() {
    let steps = [];
    let currentStepIndex = 0;

    function buildSteps(arr, target) {
        const result = [];
        result.push({
            array: [...arr],
            target: target,
            currentIndex: -1,
            status: 'initial',
            message: `Khởi tạo: Tìm kiếm tuần tự giá trị <strong>${target}</strong> trong mảng gồm ${arr.length} phần tử.`
        });

        for (let i = 0; i < arr.length; i++) {
            const val = arr[i];
            if (val === target) {
                result.push({
                    array: [...arr],
                    target: target,
                    currentIndex: i,
                    status: 'found',
                    message: `Bước ${i + 1}: Kiểm tra <strong>arr[${i}] = ${val}</strong> trùng khớp với mục tiêu <strong>${target}</strong>! Tìm thấy tại chỉ số <strong>${i}</strong>.`
                });
                return result;
            } else {
                result.push({
                    array: [...arr],
                    target: target,
                    currentIndex: i,
                    status: 'checking',
                    message: `Bước ${i + 1}: Kiểm tra <strong>arr[${i}] = ${val}</strong> &ne; <strong>${target}</strong>. Tiếp tục duyệt phần tử tiếp theo.`
                });
            }
        }

        result.push({
            array: [...arr],
            target: target,
            currentIndex: arr.length,
            status: 'not_found',
            message: `Duyệt hết mảng nhưng không tìm thấy giá trị <strong>${target}</strong>.`
        });

        return result;
    }

    function render(container, state) {
        if (!container) return;

        let html = `
            <div class="viz-array-wrapper">
                <div class="viz-target-chip">Mục tiêu tìm kiếm: <strong>${state.target}</strong></div>
                <div class="viz-array-row">
        `;

        state.array.forEach((val, idx) => {
            let cellClasses = ['viz-cell'];
            let pointerLabels = [];

            if (idx === state.currentIndex) {
                pointerLabels.push('<span class="viz-pointer-label viz-ptr-mid">i</span>');
                if (state.status === 'found') {
                    cellClasses.push('found-cell');
                } else {
                    cellClasses.push('mid-cell');
                }
            } else if (state.currentIndex > idx) {
                cellClasses.push('eliminated-cell');
            }

            html += `
                <div class="viz-cell-wrapper">
                    ${pointerLabels.join('')}
                    <div class="${cellClasses.join(' ')}">${val}</div>
                    <span class="viz-cell-index">${idx}</span>
                </div>
            `;
        });

        html += `
                </div>
            </div>
        `;

        container.innerHTML = html;
    }

    return {
        init: function(customArray, customTarget) {
            const arr = customArray && customArray.length > 0 ? customArray : [14, 29, 35, 12, 48, 67, 81];
            const target = customTarget !== undefined ? customTarget : 48;
            steps = buildSteps(arr, target);
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
