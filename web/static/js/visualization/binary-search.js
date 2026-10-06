/**
 * Binary Search Visualizer (Vanilla JS, Zero Framework)
 * Mô phỏng thuật toán Tìm kiếm Nhị phân từng bước
 */

window.BinarySearchVisualizer = (function() {
    let steps = [];
    let currentStepIndex = 0;

    function buildSteps(arr, target) {
        const result = [];
        let left = 0;
        let right = arr.length - 1;
        let stepCount = 0;

        // Trạng thái khởi tạo
        result.push({
            array: [...arr],
            target: target,
            left: left,
            right: right,
            mid: Math.floor((left + right) / 2),
            step: stepCount,
            status: 'initial',
            message: `Khởi tạo: Tìm kiếm giá trị <strong>${target}</strong> trong mảng đã sắp xếp gồm ${arr.length} phần tử.`
        });

        while (left <= right) {
            stepCount++;
            const mid = Math.floor((left + right) / 2);
            const midVal = arr[mid];

            if (midVal === target) {
                result.push({
                    array: [...arr],
                    target: target,
                    left: left,
                    right: right,
                    mid: mid,
                    step: stepCount,
                    status: 'found',
                    message: `Bước ${stepCount}: <strong>arr[${mid}] == ${midVal}</strong> trùng khớp với giá trị cần tìm <strong>${target}</strong>! Thuật toán kết thúc thành công.`
                });
                return result;
            } else if (midVal < target) {
                result.push({
                    array: [...arr],
                    target: target,
                    left: left,
                    right: right,
                    mid: mid,
                    step: stepCount,
                    status: 'searching',
                    message: `Bước ${stepCount}: Kiểm tra <strong>arr[${mid}] = ${midVal}</strong> &lt; <strong>${target}</strong>. Giá trị cần tìm nằm ở nửa phải &rarr; Cập nhật <strong>left = mid + 1 = ${mid + 1}</strong>.`
                });
                left = mid + 1;
            } else {
                result.push({
                    array: [...arr],
                    target: target,
                    left: left,
                    right: right,
                    mid: mid,
                    step: stepCount,
                    status: 'searching',
                    message: `Bước ${stepCount}: Kiểm tra <strong>arr[${mid}] = ${midVal}</strong> &gt; <strong>${target}</strong>. Giá trị cần tìm nằm ở nửa trái &rarr; Cập nhật <strong>right = mid - 1 = ${mid - 1}</strong>.`
                });
                right = mid - 1;
            }
        }

        // Không tìm thấy
        result.push({
            array: [...arr],
            target: target,
            left: left,
            right: right,
            mid: -1,
            step: stepCount + 1,
            status: 'not_found',
            message: `Kết thúc: left (${left}) &gt; right (${right}). Không tìm thấy giá trị <strong>${target}</strong> trong mảng.`
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

            if (idx === state.left) pointerLabels.push('<span class="viz-pointer-label viz-ptr-left">L</span>');
            if (idx === state.mid && state.status !== 'not_found') pointerLabels.push('<span class="viz-pointer-label viz-ptr-mid">M</span>');
            if (idx === state.right) pointerLabels.push('<span class="viz-pointer-label viz-ptr-right">R</span>');

            if (state.status === 'found' && idx === state.mid) {
                cellClasses.push('found-cell');
            } else if (idx === state.mid && state.status !== 'not_found') {
                cellClasses.push('mid-cell');
            } else if (idx >= state.left && idx <= state.right) {
                cellClasses.push('active-range');
            } else {
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
            const arr = customArray && customArray.length > 0 ? customArray : [2, 5, 8, 12, 16, 23, 38, 56, 72, 91];
            const target = customTarget !== undefined ? customTarget : 23;
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
