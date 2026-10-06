/**
 * Stack Visualizer (Vanilla JS, Zero Framework)
 * Mô phỏng Cấu trúc dữ liệu Ngăn xếp (Stack - LIFO)
 */

window.StackVisualizer = (function() {
    let steps = [];
    let currentStepIndex = 0;

    function buildSteps(operations) {
        const result = [];
        const currentStack = [];

        result.push({
            stack: [],
            action: 'init',
            val: null,
            message: 'Khởi tạo: Ngăn xếp (Stack) rỗng. Đỉnh ngăn xếp (Top) = -1.'
        });

        operations.forEach((op, idx) => {
            if (op.type === 'push') {
                currentStack.push(op.val);
                result.push({
                    stack: [...currentStack],
                    action: 'push',
                    val: op.val,
                    message: `Bước ${idx + 1}: Thực hiện <strong>push(${op.val})</strong> &rarr; Đưa phần tử ${op.val} vào đỉnh ngăn xếp. Top = ${currentStack.length - 1}.`
                });
            } else if (op.type === 'pop') {
                if (currentStack.length > 0) {
                    const popped = currentStack.pop();
                    result.push({
                        stack: [...currentStack],
                        action: 'pop',
                        val: popped,
                        message: `Bước ${idx + 1}: Thực hiện <strong>pop()</strong> &rarr; Lấy phần tử ${popped} ra khỏi đỉnh ngăn xếp. Top = ${currentStack.length - 1}.`
                    });
                }
            } else if (op.type === 'peek') {
                const topVal = currentStack.length > 0 ? currentStack[currentStack.length - 1] : null;
                result.push({
                    stack: [...currentStack],
                    action: 'peek',
                    val: topVal,
                    message: `Bước ${idx + 1}: Thực hiện <strong>peek() / top()</strong> &rarr; Xem phần tử đỉnh hiện tại là <strong>${topVal}</strong>.`
                });
            }
        });

        return result;
    }

    function render(container, state) {
        if (!container) return;

        let html = `
            <div class="viz-stack-wrapper">
                <div class="viz-stack-column">
        `;

        if (!state.stack || state.stack.length === 0) {
            html += `<div style="color:var(--text-muted); font-size:12px; text-align:center; padding: 20px 0;">(Ngăn xếp rỗng)</div>`;
        } else {
            state.stack.forEach((val, idx) => {
                const isTop = (idx === state.stack.length - 1);
                html += `
                    <div class="viz-stack-item ${isTop ? 'top-item' : ''}">
                        ${val} ${isTop ? '<span style="font-size:10px; color:var(--accent-amber); margin-left:6px;">&larr; Top</span>' : ''}
                    </div>
                `;
            });
        }

        html += `
                </div>
            </div>
        `;

        container.innerHTML = html;
    }

    return {
        init: function() {
            const defaultOps = [
                { type: 'push', val: 10 },
                { type: 'push', val: 25 },
                { type: 'push', val: 40 },
                { type: 'peek' },
                { type: 'pop' },
                { type: 'push', val: 99 },
                { type: 'pop' }
            ];
            steps = buildSteps(defaultOps);
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
