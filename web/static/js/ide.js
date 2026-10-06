/**
 * AlgoDB IDE UI Controller (Pure Vanilla JS, Zero Framework)
 * Điều phối sự kiện giao diện, soạn thảo mã, tìm kiếm hàm và kết nối engine.
 */

document.addEventListener('DOMContentLoaded', function() {
    let currentExercise = window.INITIAL_EXERCISE || null;

    // DOM Elements
    const codeEditor = document.getElementById('codeEditor');
    const lineNumbers = document.getElementById('lineNumbers');
    const outputLog = document.getElementById('outputLog');
    const execStatus = document.getElementById('execStatus');
    const btnRunCode = document.getElementById('btnRunCode');
    const btnRunTests = document.getElementById('btnRunTests');
    const btnResetCode = document.getElementById('btnResetCode');
    const btnClearConsole = document.getElementById('btnClearConsole');
    const dirtyIndicator = document.getElementById('dirtyIndicator');
    const funcSearchInput = document.getElementById('funcSearchInput');
    const functionsContainer = document.getElementById('functionsContainer');
    const allowedFuncsList = document.getElementById('allowedFuncsList');
    const problemAllowedChips = document.getElementById('problemAllowedChips');
    const testCasesContainer = document.getElementById('testCasesContainer');
    const testPassCount = document.getElementById('testPassCount');
    const sidebarResizer = document.getElementById('sidebarResizer');
    const ideSidebar = document.querySelector('.ide-sidebar');
    const btnFontDec = document.getElementById('btnFontDec');
    const btnFontInc = document.getElementById('btnFontInc');
    const fontSizeDisplay = document.getElementById('fontSizeDisplay');
    const highlighting = document.getElementById('highlighting');
    const highlightingCode = document.getElementById('highlightingCode');
    const btnStartExam = document.getElementById('btnStartExam');
    const examModeContainer = document.getElementById('examModeContainer');
    const examTimer = document.getElementById('examTimer');
    const examSuspendedOverlay = document.getElementById('examSuspendedOverlay');
    const suspendedReason = document.getElementById('suspendedReason');
    const suspendedTime = document.getElementById('suspendedTime');
    const saveStatus = document.getElementById('saveStatus');

    function escapeHtml(text) {
        if (!text) return '';
        return String(text).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    // --- Quản lý tiến độ luyện tập & Tự động lưu (Phase 7: Student Practice) ---
    function setSaveStatus(state) {
        if (!saveStatus) return;
        saveStatus.className = `save-status ${state}`;
        if (state === 'saving') {
            saveStatus.innerHTML = '&#8635; Đang lưu...';
        } else if (state === 'error') {
            saveStatus.innerHTML = '&#9888; Lỗi lưu';
        } else {
            saveStatus.innerHTML = '&#10003; Đã lưu';
        }
    }

    let autoSaveTimer = null;
    function triggerAutoSave() {
        if (!currentExercise) return;
        setSaveStatus('saving');
        if (autoSaveTimer) clearTimeout(autoSaveTimer);
        autoSaveTimer = setTimeout(function() {
            saveDraftCode(currentExercise.id, codeEditor.value);
        }, 2500); // 2.5 giây debounce
    }

    function saveDraftCode(exerciseId, code) {
        fetch('/api/practice/save', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ exercise_id: exerciseId, code: code })
        })
        .then(res => {
            if (!res.ok) throw new Error('HTTP ' + res.status);
            return res.json();
        })
        .then(data => {
            setSaveStatus('saved');
            if (data && data.status) {
                updateExerciseBullet(exerciseId, data.status);
            }
        })
        .catch(err => {
            console.error('Lỗi tự động lưu:', err);
            setSaveStatus('error');
        });
    }

    function updateExerciseBullet(exerciseId, status) {
        const bullet = document.getElementById(`exBullet-${exerciseId}`);
        if (!bullet) return;

        bullet.classList.remove('status-completed', 'status-inprogress', 'status-notstarted');
        if (status === 'completed') {
            bullet.classList.add('status-completed');
            bullet.innerHTML = '&#10003;'; // ✓
            bullet.title = 'Đã hoàn thành';
        } else if (status === 'in_progress') {
            bullet.classList.add('status-inprogress');
            bullet.innerHTML = '&#9679;'; // ●
            bullet.title = 'Đang làm';
        } else {
            bullet.classList.add('status-notstarted');
            bullet.innerHTML = '&#9675;'; // ○
            bullet.title = 'Chưa làm';
        }
    }

    function loadAllProgressStates() {
        fetch('/api/practice/all-states')
            .then(res => {
                if (!res.ok) return [];
                return res.json();
            })
            .then(states => {
                if (Array.isArray(states)) {
                    states.forEach(st => {
                        updateExerciseBullet(st.exercise_id, st.status);
                    });
                }
            })
            .catch(err => {
                console.error('Lỗi nạp trạng thái tiến độ:', err);
            });
    }

    function loadSavedDraft(exerciseId) {
        fetch(`/api/practice/state?exercise_id=${exerciseId}`)
            .then(res => {
                if (!res.ok) return null;
                return res.json();
            })
            .then(state => {
                if (state && state.last_code && state.last_code.trim() !== '') {
                    codeEditor.value = state.last_code;
                    updateLineNumbers();
                    updateHighlighting();
                    const initial = currentExercise.initialCode || currentExercise.initial_code || '';
                    if (state.last_code !== initial) {
                        dirtyIndicator.style.display = 'inline';
                    } else {
                        dirtyIndicator.style.display = 'none';
                    }
                }
                if (state && state.status) {
                    updateExerciseBullet(exerciseId, state.status);
                }
                setSaveStatus('saved');
            })
            .catch(err => {
                console.error('Lỗi lấy tiến độ bài tập:', err);
            });
    }

    function submitPracticeResult(exerciseId, code, score, passed) {
        fetch('/api/practice/submit', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                exercise_id: exerciseId,
                code: code,
                score: score,
                passed: passed
            })
        })
        .then(res => {
            if (!res.ok) throw new Error('HTTP ' + res.status);
            return res.json();
        })
        .then(data => {
            if (data && data.status) {
                updateExerciseBullet(exerciseId, data.status);
            }
        })
        .catch(err => {
            console.error('Lỗi ghi nhận kết quả bài tập:', err);
        });
    }

    // Ghi nhận hành động học tập (run, test, hint) - Phase 8: Attempt Metrics
    function recordAttemptAction(action) {
        if (!currentExercise) return;
        fetch('/api/attempt/action', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ exercise_id: currentExercise.id, action: action })
        }).catch(err => console.error('Lỗi ghi nhận action metric:', err));
    }

    // Lưu bản ghi lịch sử nộp bài chính thức & nhận kết quả Server-Side Judge (Phase 10)
    function createSubmission(exerciseId, code, score, passedTests, totalTests) {
        fetch('/api/submissions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                exercise_id: exerciseId,
                source_code: code,
                score: score,
                passed_tests: passedTests,
                total_tests: totalTests
            })
        })
        .then(res => res.json())
        .then(data => {
            if (data && data.judge) {
                console.log('[Server Judge] Điểm chính thức:', data.judge.score, 'Đạt:', data.judge.passed_tests + '/' + data.judge.total_tests);
            }
        })
        .catch(err => console.error('Lỗi lưu submission:', err));
    }

    // Bộ phân tích cú pháp và tô màu mã Python (Python Syntax Highlighter)
    function highlightPython(code) {
        if (!code) return '';

        const tokenRegex = /("""[\s\S]*?"""|'''[\s\S]*?'''|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*')|(#.*$)|(\b\d+(?:\.\d+)?\b)|(\b(?:def|return|if|elif|else|for|while|in|is|not|and|or|import|from|class|try|except|finally|raise|pass|break|continue|lambda|as|with|yield|async|await)\b)|(\b(?:True|False|None|self|cls|int|float|str|bool|list|dict|set|tuple)\b)|(\b[a-zA-Z_]\w*(?=\s*\())/gm;

        let lastIdx = 0;
        let html = '';

        code.replace(tokenRegex, function(match, str, com, num, kw, builtin, fn, offset) {
            if (offset > lastIdx) {
                html += escapeHtml(code.slice(lastIdx, offset));
            }

            if (str) {
                html += `<span class="token-str">${escapeHtml(str)}</span>`;
            } else if (com) {
                html += `<span class="token-com">${escapeHtml(com)}</span>`;
            } else if (num) {
                html += `<span class="token-num">${escapeHtml(num)}</span>`;
            } else if (kw) {
                html += `<span class="token-kw">${escapeHtml(kw)}</span>`;
            } else if (builtin) {
                html += `<span class="token-builtin">${escapeHtml(builtin)}</span>`;
            } else if (fn) {
                html += `<span class="token-fn">${escapeHtml(fn)}</span>`;
            } else {
                html += escapeHtml(match);
            }

            lastIdx = offset + match.length;
            return match;
        });

        if (lastIdx < code.length) {
            html += escapeHtml(code.slice(lastIdx));
        }

        return html;
    }

    function updateHighlighting() {
        if (highlightingCode) {
            highlightingCode.innerHTML = highlightPython(codeEditor.value) + '\n';
        }
    }

    // 1. Khởi tạo Editor & Line Numbers
    function updateLineNumbers() {
        const lines = codeEditor.value.split('\n').length;
        let lineNumbersHtml = '';
        for (let i = 1; i <= lines; i++) {
            lineNumbersHtml += i + '<br>';
        }
        lineNumbers.innerHTML = lineNumbersHtml;
    }

    codeEditor.addEventListener('input', function() {
        updateLineNumbers();
        updateHighlighting();
        if (currentExercise && codeEditor.value !== currentExercise.initialCode) {
            dirtyIndicator.style.display = 'inline';
        } else {
            dirtyIndicator.style.display = 'none';
        }
        triggerAutoSave();
    });

    codeEditor.addEventListener('scroll', function() {
        lineNumbers.scrollTop = codeEditor.scrollTop;
        if (highlighting) {
            highlighting.scrollTop = codeEditor.scrollTop;
            highlighting.scrollLeft = codeEditor.scrollLeft;
        }
    });

    // Xử lý phím Tab (4 dấu cách) và phím Enter (giữ lề)
    codeEditor.addEventListener('keydown', function(e) {
        if (e.key === 'Tab') {
            e.preventDefault();
            const start = this.selectionStart;
            const end = this.selectionEnd;
            this.value = this.value.substring(0, start) + '    ' + this.value.substring(end);
            this.selectionStart = this.selectionEnd = start + 4;
            updateLineNumbers();
            updateHighlighting();
            triggerAutoSave();
        } else if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
            e.preventDefault();
            runCode();
        } else if (e.key === 'Enter') {
            // Tự động căn lề (auto-indentation)
            const start = this.selectionStart;
            const lines = this.value.substring(0, start).split('\n');
            const currentLine = lines[lines.length - 1];
            const match = currentLine.match(/^(\s+)/);
            let indent = match ? match[1] : '';
            if (currentLine.trim().endsWith(':')) {
                indent += '    ';
            }
            if (indent) {
                e.preventDefault();
                this.value = this.value.substring(0, start) + '\n' + indent + this.value.substring(this.selectionEnd);
                this.selectionStart = this.selectionEnd = start + 1 + indent.length;
                updateLineNumbers();
                updateHighlighting();
                triggerAutoSave();
            }
        }
    });

    // 2. Chuyển đổi Tabs Sidebar
    document.querySelectorAll('.sidebar-tab').forEach(tab => {
        tab.addEventListener('click', function() {
            document.querySelectorAll('.sidebar-tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

            this.classList.add('active');
            const targetId = this.getAttribute('data-tab');
            const targetContent = document.getElementById(targetId);
            if (targetContent) targetContent.classList.add('active');

            if (targetId === 'tab-problem') {
                recordAttemptAction('hint');
            }
        });
    });

    // 3. Chuyển đổi Tabs Console
    document.querySelectorAll('.console-tab').forEach(tab => {
        tab.addEventListener('click', function() {
            document.querySelectorAll('.console-tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.console-tab-view').forEach(v => v.classList.remove('active'));

            this.classList.add('active');
            const targetId = this.getAttribute('data-target');
            const targetView = document.getElementById(targetId);
            if (targetView) targetView.classList.add('active');
        });
    });

    // 4. Tìm kiếm hàm theo thời gian thực
    if (funcSearchInput) {
        funcSearchInput.addEventListener('input', function() {
            const query = this.value.toLowerCase().trim();
            const cards = functionsContainer.querySelectorAll('.function-card');
            cards.forEach(card => {
                const name = card.getAttribute('data-name').toLowerCase();
                const cat = card.getAttribute('data-category').toLowerCase();
                const text = card.textContent.toLowerCase();
                if (name.includes(query) || cat.includes(query) || text.includes(query)) {
                    card.style.display = 'block';
                } else {
                    card.style.display = 'none';
                }
            });
        });
    }

    // 5. Cập nhật giao diện bài tập hiện tại
    window.renderExerciseDetails = function(ex) {
        currentExercise = ex;
        if (!ex) return;

        // Cập nhật header
        const topTitle = document.getElementById('topExerciseTitle');
        const topDiff = document.getElementById('topExerciseDiff');
        if (topTitle) topTitle.textContent = ex.title;
        if (topDiff) {
            topDiff.textContent = ex.difficulty;
            topDiff.className = `badge-difficulty badge-${ex.difficulty}`;
        }

        // Cập nhật tab Đề bài (nếu mở tab Đề bài)
        const problemTitle = document.getElementById('problemTitle');
        const problemDesc = document.getElementById('problemDesc');
        const problemHint = document.getElementById('problemHint');
        if (problemTitle) problemTitle.textContent = ex.title;
        if (problemDesc) problemDesc.textContent = ex.description;
        if (problemHint) problemHint.textContent = ex.hint || ex.solution_hint || 'Không có gợi ý.';

        // Cập nhật Allowed functions chips
        const allowed = ex.allowedFunctions || ex.allowed_functions || [];
        if (problemAllowedChips) {
            problemAllowedChips.innerHTML = allowed.map(fn => `<span class="func-chip">${escapeHtml(fn)}()</span>`).join('');
        }
        if (allowedFuncsList) {
            allowedFuncsList.innerHTML = allowed.length > 0
                ? `Bài này chỉ được phép dùng: <strong>${allowed.join(', ')}</strong>`
                : 'Bài tập cho phép sử dụng các hàm cơ bản.';
        }

        // Cập nhật khung đề bài nội tuyến (Inline Problem Drawer) ngay dưới bài tập được chọn
        document.querySelectorAll('.inline-problem-panel').forEach(p => {
            p.classList.add('collapsed');
        });
        document.querySelectorAll('.btn-toggle-problem .toggle-text').forEach(t => {
            t.textContent = 'Xem đề';
        });

        const activeDrawer = document.getElementById(`inlineProblem-${ex.id}`);
        const activeItem = document.getElementById(`exercise-item-${ex.id}`);
        if (activeDrawer && activeItem) {
            let chipsHtml = '';
            if (allowed.length > 0) {
                chipsHtml = `
                    <div class="inline-section-title">Hàm cho phép:</div>
                    <div class="chip-container">
                        ${allowed.map(fn => `<span class="func-chip">${escapeHtml(fn)}()</span>`).join('')}
                    </div>
                `;
            }

            let hintHtml = '';
            const hint = ex.hint || ex.solution_hint;
            if (hint) {
                hintHtml = `<div class="inline-hint-box"><strong>Gợi ý:</strong> ${escapeHtml(hint)}</div>`;
            }

            activeDrawer.innerHTML = `
                <div class="inline-problem-desc">${escapeHtml(ex.description)}</div>
                ${chipsHtml}
                ${hintHtml}
            `;
            activeDrawer.classList.remove('collapsed');

            const btn = activeItem.querySelector('.btn-toggle-problem');
            if (btn) {
                btn.querySelector('.toggle-text').textContent = 'Ẩn đề';
            }
        }

        // Đặt lại mã nguồn
        codeEditor.value = ex.initialCode || ex.initial_code || '';
        dirtyIndicator.style.display = 'none';
        updateLineNumbers();
        updateHighlighting();

        // Xử lý bài tập trắc nghiệm / Code Tracing (Phase 13)
        const quizSection = document.getElementById('quizSection');
        const editorBody = document.getElementById('editorBody');
        const isQuiz = (ex.exercise_type === 'multiple_choice' || ex.exercise_type === 'quiz');

        if (quizSection) {
            if (isQuiz) {
                quizSection.style.display = 'block';
                const qTitle = document.getElementById('quizTitle');
                const qDesc = document.getElementById('quizQuestionDesc');
                const qMsg = document.getElementById('quizResultMsg');
                if (qTitle) qTitle.textContent = ex.title;
                if (qDesc) qDesc.textContent = ex.description;
                if (qMsg) {
                    qMsg.textContent = '';
                    qMsg.className = 'quiz-result-msg';
                }
                loadQuizOptions(ex.id);
            } else {
                quizSection.style.display = 'none';
            }
        }

        if (editorBody) {
            const hasInitialCode = (ex.initialCode || ex.initial_code || '').trim().length > 0;
            if (isQuiz && !hasInitialCode) {
                editorBody.style.display = 'none';
            } else {
                editorBody.style.display = 'flex';
            }
        }

        // Nạp bản nháp đã lưu của học viên từ server (nếu có)
        loadSavedDraft(ex.id);
    };

    // Phase 13: Tải các lựa chọn trắc nghiệm từ API bảo mật (tuyệt đối không lộ đáp án đúng)
    function loadQuizOptions(exerciseId) {
        const listEl = document.getElementById('quizOptionsList');
        if (listEl) {
            listEl.innerHTML = '<div style="color:var(--text-muted); font-size:13px;">Đang tải các phương án lựa chọn...</div>';
        }
        fetch(`/api/quiz?exercise_id=${exerciseId}`)
            .then(res => {
                if (!res.ok) throw new Error('HTTP ' + res.status);
                return res.json();
            })
            .then(options => {
                renderQuizOptions(options);
            })
            .catch(err => {
                if (listEl) {
                    listEl.innerHTML = `<div style="color:var(--accent-ruby); font-size:13px;">Lỗi tải phương án: ${escapeHtml(err.message)}</div>`;
                }
            });
    }

    function renderQuizOptions(options) {
        const listEl = document.getElementById('quizOptionsList');
        if (!listEl) return;
        if (!options || options.length === 0) {
            listEl.innerHTML = '<div style="color:var(--text-muted); font-size:13px;">Chưa có phương án nào được cấu hình cho câu hỏi này.</div>';
            return;
        }
        listEl.innerHTML = options.map((opt, idx) => {
            const charLabel = String.fromCharCode(65 + idx);
            return `
                <label class="quiz-option-label" id="quizLabel-${opt.id}">
                    <input type="radio" name="quizOption" value="${opt.id}" class="quiz-option-radio" onchange="handleQuizOptionSelect(${opt.id})">
                    <strong style="color:var(--accent-green); min-width:20px;">${charLabel}.</strong>
                    <span class="quiz-option-text">${escapeHtml(opt.content)}</span>
                </label>
            `;
        }).join('');
    }

    window.handleQuizOptionSelect = function(selectedId) {
        document.querySelectorAll('.quiz-option-label').forEach(lbl => {
            lbl.classList.remove('selected');
        });
        const activeLbl = document.getElementById(`quizLabel-${selectedId}`);
        if (activeLbl) activeLbl.classList.add('selected');
    };

    window.submitQuizChoice = function() {
        if (!currentExercise) return;
        const selectedRadio = document.querySelector('input[name="quizOption"]:checked');
        const msgEl = document.getElementById('quizResultMsg');
        if (!selectedRadio) {
            if (msgEl) {
                msgEl.textContent = 'Vui lòng chọn một đáp án trước khi nộp!';
                msgEl.className = 'quiz-result-msg incorrect';
            }
            return;
        }
        const optionId = parseInt(selectedRadio.value, 10);
        const btnSubmit = document.getElementById('btnSubmitQuiz');
        if (btnSubmit) btnSubmit.disabled = true;

        fetch('/api/quiz/submit', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ exercise_id: currentExercise.id, option_id: optionId })
        })
        .then(res => {
            if (!res.ok) throw new Error('HTTP ' + res.status);
            return res.json();
        })
        .then(data => {
            if (btnSubmit) btnSubmit.disabled = false;
            if (msgEl) {
                msgEl.textContent = data.message || (data.is_correct ? 'Chính xác! Điểm: ' + data.score : 'Chưa chính xác!');
                msgEl.className = 'quiz-result-msg ' + (data.is_correct ? 'correct' : 'incorrect');
            }
            if (data.is_correct) {
                updateExerciseBullet(currentExercise.id, 'completed');
            } else {
                updateExerciseBullet(currentExercise.id, 'in_progress');
            }
        })
        .catch(err => {
            if (btnSubmit) btnSubmit.disabled = false;
            if (msgEl) {
                msgEl.textContent = 'Lỗi nộp bài: ' + err.message;
                msgEl.className = 'quiz-result-msg incorrect';
            }
        });
    };

    // Chuyển đổi trạng thái Ẩn/Hiện đề bài nội tuyến
    window.toggleProblemDrawer = function(event, id) {
        if (event) event.stopPropagation();
        const drawer = document.getElementById(`inlineProblem-${id}`);
        const item = document.getElementById(`exercise-item-${id}`);
        if (!drawer || !item) return;

        const btn = item.querySelector('.btn-toggle-problem');
        const isCollapsed = drawer.classList.toggle('collapsed');

        if (btn) {
            btn.querySelector('.toggle-text').textContent = isCollapsed ? 'Xem đề' : 'Ẩn đề';
        }
        if (!isCollapsed) {
            recordAttemptAction('hint');
        }
    };

    // 6. Tải bài tập khi người dùng bấm chọn
    window.loadExercise = function(id) {
        fetch(`/api/exercise?id=${id}`)
            .then(res => res.json())
            .then(ex => {
                // Cập nhật danh sách active
                document.querySelectorAll('.exercise-item').forEach(item => {
                    item.classList.remove('active');
                    if (parseInt(item.getAttribute('data-id')) === id) {
                        item.classList.add('active');
                    }
                });
                renderExerciseDetails(ex);
            })
            .catch(err => {
                console.error('Lỗi khi nạp bài tập:', err);
            });
    };

    // 7. Chạy mã Python
    function runCode() {
        recordAttemptAction('run');
        const code = codeEditor.value;
        const allowed = currentExercise ? (currentExercise.allowedFunctions || currentExercise.allowed_functions || []) : [];

        // Chuyển sang tab console
        document.querySelector('.console-tab[data-target="consoleOutput"]').click();

        // Xóa console cũ trước khi thực thi
        outputLog.textContent = '';
        execStatus.textContent = "Đang chạy...";
        execStatus.style.color = "var(--accent-amber)";

        // Bước 1: Kiểm tra Whitelist giới hạn hàm
        const check = PythonEngine.validateCode(code, allowed);
        if (!check.isValid) {
            outputLog.innerHTML = `<span class="log-stderr">${check.errorMsg}</span>\n`;
            execStatus.textContent = "Bị chặn bởi quy tắc";
            execStatus.style.color = "var(--accent-red)";
            return;
        }

        // Bước 2: Thực thi chương trình
        PythonEngine.runPython(
            code,
            function(out) {
                // Stream output trực tiếp vào console
                outputLog.appendChild(document.createTextNode(out));
                outputLog.scrollTop = outputLog.scrollHeight;
            },
            function(err) {
                const errSpan = document.createElement('span');
                errSpan.className = 'log-stderr';
                errSpan.textContent = `\nTraceback (most recent call last):\n  ${err}\n`;
                outputLog.appendChild(errSpan);
            },
            function(success, duration) {
                const footer = document.createElement('span');
                footer.className = success ? 'log-system' : 'log-stderr';
                footer.textContent = `\n--------------------------------\nProcess finished with exit code ${success ? '0' : '1'} (${duration}ms)\n`;
                outputLog.appendChild(footer);

                execStatus.textContent = success ? "Hoàn tất" : "Lỗi";
                execStatus.style.color = success ? "var(--accent-emerald)" : "var(--accent-red)";
            }
        );
    }

    // 8. Chấm thử (Test Cases)
    function runTests() {
        if (currentExercise && (currentExercise.exercise_type === 'multiple_choice' || currentExercise.exercise_type === 'quiz')) {
            submitQuizChoice();
            return;
        }

        recordAttemptAction('test');
        const code = codeEditor.value;
        const allowed = currentExercise ? (currentExercise.allowedFunctions || currentExercise.allowed_functions || []) : [];

        // Mở tab Test Cases
        document.querySelector('.console-tab[data-target="testResults"]').click();

        // Kiểm tra whitelist
        const check = PythonEngine.validateCode(code, allowed);
        if (!check.isValid) {
            testCasesContainer.innerHTML = `<div class="testcase-item fail">
                <div class="tc-header tc-fail">Quy tắc không hợp lệ</div>
                <div class="tc-detail">${check.errorMsg}</div>
            </div>`;
            return;
        }

        let testCases = [];
        try {
            const raw = currentExercise.testCasesJSON || currentExercise.test_cases_json || currentExercise.test_cases || "[]";
            testCases = typeof raw === 'string' ? JSON.parse(raw) : raw;
        } catch (e) {
            testCases = [];
        }

        testCasesContainer.innerHTML = `<div class="empty-state">Đang chạy kiểm thử qua các test cases...</div>`;

        PythonEngine.runTests(
            code,
            testCases,
            null,
            function(results, passCount, totalCount) {
                testPassCount.textContent = `${passCount}/${totalCount}`;
                testPassCount.style.color = (passCount === totalCount && totalCount > 0) ? "var(--accent-emerald)" : "var(--accent-amber)";

                if (results.length === 0) {
                    testCasesContainer.innerHTML = `<div class="empty-state">Bài này không có test cases định sẵn.</div>`;
                    return;
                }

                testCasesContainer.innerHTML = results.map(r => `
                    <div class="testcase-item ${r.passed ? 'pass' : 'fail'}">
                        <div class="tc-header">
                            <span>Test Case #${r.index}</span>
                            <span class="${r.passed ? 'tc-pass' : 'tc-fail'}">${r.passed ? 'ĐẠT (PASS)' : 'KHÔNG ĐẠT (FAIL)'}</span>
                        </div>
                        <div class="tc-detail"><strong>Đầu vào:</strong> <code>${escapeHtml(r.input)}</code></div>
                        <div class="tc-detail"><strong>Kỳ vọng:</strong> <code>${escapeHtml(r.expected)}</code></div>
                        <div class="tc-detail"><strong>Thực tế trả về:</strong> <code class="${r.passed ? 'tc-pass' : 'tc-fail'}">${escapeHtml(r.actual)}</code></div>
                    </div>
                `).join('');

                // Ghi nhận kết quả luyện tập vào cơ sở dữ liệu (Phase 7: Practice & Phase 8: Submission)
                if (currentExercise && totalCount > 0) {
                    const passed = (passCount === totalCount);
                    const score = (passCount / totalCount) * 100;
                    submitPracticeResult(currentExercise.id, code, score, passed);
                    createSubmission(currentExercise.id, code, score, passCount, totalCount);
                }
            }
        );
    }

    // Event Listeners
    btnRunCode.addEventListener('click', runCode);
    btnRunTests.addEventListener('click', runTests);

    btnResetCode.addEventListener('click', function() {
        if (currentExercise) {
            if (confirm("Bạn có chắc chắn muốn đặt lại mã nguồn về trạng thái ban đầu?")) {
                const initial = currentExercise.initialCode || currentExercise.initial_code || '';
                codeEditor.value = initial;
                updateLineNumbers();
                updateHighlighting();
                dirtyIndicator.style.display = 'none';
                saveDraftCode(currentExercise.id, initial);
            }
        }
    });

    btnClearConsole.addEventListener('click', function() {
        outputLog.textContent = '';
        execStatus.textContent = "Sẵn sàng";
        execStatus.style.color = "var(--accent-emerald)";
    });

    // 9. Kéo co giãn thanh bên trái (Resizable Sidebar)
    if (sidebarResizer && ideSidebar) {
        // Phục hồi kích thước đã lưu
        const savedWidth = localStorage.getItem('algodb_sidebar_width');
        if (savedWidth) {
            ideSidebar.style.width = savedWidth;
        }

        let isResizing = false;
        let startX = 0;
        let startWidth = 0;

        sidebarResizer.addEventListener('mousedown', function(e) {
            isResizing = true;
            startX = e.clientX;
            startWidth = ideSidebar.getBoundingClientRect().width;
            sidebarResizer.classList.add('is-resizing');
            document.body.classList.add('is-resizing');
            e.preventDefault();
        });

        document.addEventListener('mousemove', function(e) {
            if (!isResizing) return;
            const deltaX = e.clientX - startX;
            let newWidth = startWidth + deltaX;

            const minWidth = 220;
            const maxWidth = Math.round(window.innerWidth * 0.7);

            if (newWidth < minWidth) newWidth = minWidth;
            if (newWidth > maxWidth) newWidth = maxWidth;

            ideSidebar.style.width = newWidth + 'px';
        });

        document.addEventListener('mouseup', function() {
            if (isResizing) {
                isResizing = false;
                sidebarResizer.classList.remove('is-resizing');
                document.body.classList.remove('is-resizing');
                localStorage.setItem('algodb_sidebar_width', ideSidebar.style.width);
            }
        });
    }

    // 10. Tăng/giảm cỡ chữ hiển thị (Font Size Controls)
    let currentFontSize = parseInt(localStorage.getItem('algodb_font_size'), 10) || 13;

    function applyFontSize(size) {
        if (size < 11) size = 11;
        if (size > 24) size = 24;
        currentFontSize = size;

        const lineHeight = Math.round(size * 1.54);
        codeEditor.style.fontSize = size + 'px';
        codeEditor.style.lineHeight = lineHeight + 'px';
        lineNumbers.style.fontSize = size + 'px';
        lineNumbers.style.lineHeight = lineHeight + 'px';
        if (highlighting) {
            highlighting.style.fontSize = size + 'px';
            highlighting.style.lineHeight = lineHeight + 'px';
        }

        if (outputLog) {
            outputLog.style.fontSize = Math.max(11, size - 1) + 'px';
            outputLog.style.lineHeight = Math.max(16, lineHeight - 2) + 'px';
        }

        if (fontSizeDisplay) {
            fontSizeDisplay.textContent = size + 'px';
        }

        localStorage.setItem('algodb_font_size', size);
        updateLineNumbers();
    }

    if (btnFontDec) {
        btnFontDec.addEventListener('click', function() {
            applyFontSize(currentFontSize - 1);
        });
    }

    if (btnFontInc) {
        btnFontInc.addEventListener('click', function() {
            applyFontSize(currentFontSize + 1);
        });
    }

    // Áp dụng cỡ chữ ban đầu
    applyFontSize(currentFontSize);

    // ==========================================================================
    // 11. Chế độ thi & Chống gian lận (Exam Mode & Anti-Cheating)
    // ==========================================================================
    let isExamMode = false;
    let isExamSuspended = false;
    let antiCheatMonitoring = false;
    let examTimerInterval = null;
    let examSecondsRemaining = 45 * 60; // Mặc định thời gian làm bài: 45 phút

    function formatTime(totalSeconds) {
        const mins = Math.floor(totalSeconds / 60);
        const secs = totalSeconds % 60;
        return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
    }

    function enterFullscreen() {
        const el = document.documentElement;
        if (el.requestFullscreen) {
            return el.requestFullscreen();
        } else if (el.webkitRequestFullscreen) {
            return el.webkitRequestFullscreen();
        } else if (el.mozRequestFullScreen) {
            return el.mozRequestFullScreen();
        } else if (el.msRequestFullscreen) {
            return el.msRequestFullscreen();
        }
        return Promise.resolve();
    }

    function suspendExam(reason) {
        if (!isExamMode || isExamSuspended) return;
        isExamSuspended = true;
        antiCheatMonitoring = false;

        if (examTimerInterval) {
            clearInterval(examTimerInterval);
            examTimerInterval = null;
        }

        // Khóa hoàn toàn trình soạn thảo
        codeEditor.readOnly = true;
        codeEditor.disabled = true;

        // Vô hiệu hóa toàn bộ nút thao tác
        if (btnRunCode) btnRunCode.disabled = true;
        if (btnRunTests) btnRunTests.disabled = true;
        if (btnResetCode) btnResetCode.disabled = true;
        if (btnStartExam) {
            btnStartExam.disabled = true;
            btnStartExam.innerHTML = '<span class="icon">&#9888;</span> Bị đình chỉ';
        }

        // Cập nhật trạng thái và terminal console
        if (execStatus) {
            execStatus.textContent = "Đình chỉ thi (Gian lận)";
            execStatus.style.color = "var(--accent-red)";
        }

        const now = new Date();
        const timeStr = now.toLocaleTimeString('vi-VN') + ' ngày ' + now.toLocaleDateString('vi-VN');

        if (outputLog) {
            const warnBlock = document.createElement('div');
            warnBlock.className = 'log-stderr';
            warnBlock.style.padding = '10px';
            warnBlock.style.border = '1px solid #ef4444';
            warnBlock.style.backgroundColor = 'rgba(239, 68, 68, 0.1)';
            warnBlock.style.marginTop = '10px';
            warnBlock.innerHTML = `\n[CẢNH BÁO HỆ THỐNG]\nBÀI THI ĐÃ BỊ ĐÌNH CHỈ DO GIAN LẬN!\n- Lý do vi phạm: ${escapeHtml(reason)}\n- Thời điểm: ${timeStr}\n- Toàn bộ thao tác chạy mã, chấm điểm và sửa đổi đã bị vô hiệu hóa vĩnh viễn.\n`;
            outputLog.appendChild(warnBlock);
            outputLog.scrollTop = outputLog.scrollHeight;
        }

        // Kích hoạt overlay cảnh báo toàn màn hình
        if (suspendedReason) {
            suspendedReason.textContent = reason;
        }
        if (suspendedTime) {
            suspendedTime.textContent = 'Thời điểm vi phạm: ' + timeStr;
        }
        if (examSuspendedOverlay) {
            examSuspendedOverlay.style.display = 'flex';
        }
    }

    function startExamMode() {
        const confirmMsg = "QUY CHẾ PHÒNG THI TRỰC TUYẾN:\n\n" +
            "1. Hệ thống sẽ tự động chuyển sang chế độ TOÀN MÀN HÌNH (Fullscreen).\n" +
            "2. Tuyệt đối KHÔNG SAO CHÉP / DÁN (Copy/Paste), không mở chuột phải hoặc dùng DevTools.\n" +
            "3. Tuyệt đối KHÔNG CHUYỂN TAB, KHÔNG THOÁT TOÀN MÀN HÌNH HOẶC RỜI CỬA SỔ LÀM BÀI.\n" +
            "Mọi hành vi mất tiêu điểm/nhảy trang sẽ lập tức bị hệ thống XÁC ĐỊNH LÀ GIAN LẬN và ĐÌNH CHỈ THI.\n\n" +
            "Bạn có chắc chắn muốn bắt đầu làm bài thi ngay bây giờ?";

        if (!confirm(confirmMsg)) return;

        isExamMode = true;
        isExamSuspended = false;

        // Bật toàn màn hình
        enterFullscreen().catch(err => {
            console.warn("Fullscreen request:", err);
        });

        document.body.classList.add('exam-mode-active');

        if (btnStartExam) {
            btnStartExam.disabled = true;
            btnStartExam.innerHTML = '<span class="icon">&#128274;</span> Đang trong kỳ thi';
        }

        if (examModeContainer) {
            examModeContainer.style.display = 'flex';
        }

        if (examTimer) {
            examTimer.textContent = formatTime(examSecondsRemaining);
        }

        // Bắt đầu đếm ngược thời gian làm bài
        examTimerInterval = setInterval(function() {
            examSecondsRemaining--;

            if (examTimer) {
                examTimer.textContent = formatTime(examSecondsRemaining);
                if (examSecondsRemaining <= 300) {
                    examTimer.classList.add('urgent');
                }
            }

            if (examSecondsRemaining <= 0) {
                clearInterval(examTimerInterval);
                examTimerInterval = null;
                codeEditor.readOnly = true;
                if (btnRunCode) btnRunCode.disabled = true;
                if (btnRunTests) btnRunTests.disabled = true;
                alert("HẾT GIỜ LÀM BÀI! Hệ thống đã tự động khóa bài thi.");
            }
        }, 1000);

        // Kích hoạt giám sát chống gian lận sau 1000ms để trình duyệt chuyển toàn màn hình ổn định
        setTimeout(function() {
            antiCheatMonitoring = true;
        }, 1000);
    }

    if (btnStartExam) {
        btnStartExam.addEventListener('click', startExamMode);
    }

    // 1. Chặn Copy, Cut, Paste, Chuột phải
    ['copy', 'cut', 'paste', 'contextmenu'].forEach(eventType => {
        window.addEventListener(eventType, function(e) {
            if (isExamMode && !isExamSuspended) {
                e.preventDefault();
                e.stopPropagation();
            }
        }, true);
    });

    // 2. Chặn các phím tắt gian lận (Ctrl+C, Ctrl+V, Ctrl+X, Ctrl+U, F12, Ctrl+Shift+I/J/C)
    window.addEventListener('keydown', function(e) {
        if (!isExamMode || isExamSuspended) return;

        if (e.key === 'F12') {
            e.preventDefault();
            e.stopPropagation();
            return;
        }

        const isCtrl = e.ctrlKey || e.metaKey;
        if (isCtrl) {
            const k = (e.key || '').toLowerCase();
            // Cấm Copy, Paste, Cut, View Source (U), Print (P), Save (S)
            if (['c', 'v', 'x', 'u', 'p', 's'].includes(k)) {
                e.preventDefault();
                e.stopPropagation();
                return;
            }

            // Cấm mở Developer Tools: Ctrl + Shift + I/J/C
            if (e.shiftKey && ['i', 'j', 'c'].includes(k)) {
                e.preventDefault();
                e.stopPropagation();
                return;
            }
        }
    }, true);

    // 3. Phát hiện chuyển tab hoặc rời trang (visibilitychange)
    document.addEventListener('visibilitychange', function() {
        if (isExamMode && !isExamSuspended && antiCheatMonitoring) {
            if (document.hidden || document.visibilityState === 'hidden') {
                suspendExam('Phát hiện rời khỏi trang thi hoặc chuyển sang tab khác (Visibility Hidden).');
            }
        }
    });

    // 4. Phát hiện mất tiêu điểm màn hình (blur / mất focus cửa sổ)
    window.addEventListener('blur', function() {
        if (isExamMode && !isExamSuspended && antiCheatMonitoring) {
            suspendExam('Phát hiện mất tiêu điểm màn hình làm bài / chuyển sang ứng dụng khác (Window Blur).');
        }
    });

    // 5. Phát hiện thoát khỏi chế độ toàn màn hình
    const handleFullscreenChange = function() {
        if (isExamMode && !isExamSuspended && antiCheatMonitoring) {
            const isFullscreen = document.fullscreenElement ||
                document.webkitFullscreenElement ||
                document.mozFullScreenElement ||
                document.msFullscreenElement;
            if (!isFullscreen) {
                suspendExam('Phát hiện hành vi thoát khỏi chế độ toàn màn hình.');
            }
        }
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
    document.addEventListener('mozfullscreenchange', handleFullscreenChange);
    document.addEventListener('MSFullscreenChange', handleFullscreenChange);

    // Khởi tạo bài tập đầu tiên
    if (window.INITIAL_EXERCISE) {
        renderExerciseDetails(window.INITIAL_EXERCISE);
    }
    updateLineNumbers();

    // Tải toàn bộ trạng thái tiến độ luyện tập cho Sidebar (Phase 7)
    loadAllProgressStates();
});
