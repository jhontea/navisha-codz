/**
 * Main Application JavaScript
 * Coding Challenge Website
 * 
 * Handles routing, problem list filtering, problem detail rendering,
 * code submission, and UI state management.
 */

(function() {
  'use strict';

  // ---- API Base URL ----
  const API_BASE = '/api';

  // ---- Application State ----
  const state = {
    problems: [],
    filteredProblems: [],
    currentProblem: null,
    categories: [],
    difficulties: ['all', 'easy', 'medium', 'hard', 'expert'],
    filters: {
      difficulty: 'all',
      category: 'all',
      search: ''
    },
    editor: null,
    hints: [],
    revealedHintCount: 0,
    isSubmitting: false,
    testResults: null
  };

  // ---- DOM Elements ----
  const elements = {};

  function cacheElements() {
    elements.problemGrid = document.getElementById('problem-grid');
    elements.filterDifficulty = document.getElementById('filter-difficulty');
    elements.filterCategory = document.getElementById('filter-category');
    elements.filterSearch = document.getElementById('filter-search');
    elements.loadingState = document.getElementById('loading-state');
    elements.emptyState = document.getElementById('empty-state');
    elements.sidebar = document.getElementById('sidebar');
    elements.sidebarToggle = document.getElementById('sidebar-toggle');
    elements.sidebarOverlay = document.getElementById('sidebar-overlay');
    elements.pageTitle = document.getElementById('page-title');
    elements.backLink = document.getElementById('back-link');
    elements.viewList = document.getElementById('view-list');
    elements.viewDetail = document.getElementById('view-detail');
    elements.problemInfo = document.getElementById('problem-info');
    elements.testResults = document.getElementById('test-results');
    elements.btnSubmit = document.getElementById('btn-submit');
    elements.btnReset = document.getElementById('btn-reset');
    elements.healthStatus = document.getElementById('health-status');
    elements.confirmModal = document.getElementById('confirm-modal');
    elements.modalTitle = document.getElementById('modal-title');
    elements.modalBody = document.getElementById('modal-body');
    elements.modalCancel = document.getElementById('modal-cancel');
    elements.modalConfirm = document.getElementById('modal-confirm');
    elements.errorState = document.getElementById('error-state');
    elements.errorTitle = document.getElementById('error-title');
    elements.errorMessage = document.getElementById('error-message');
    elements.problemDetail = document.getElementById('problem-detail');
    elements.loadingStateDetail = document.getElementById('loading-state');
  }

  // ---- API Functions ----
  async function fetchProblems(params = {}) {
    const queryParams = new URLSearchParams();
    if (params.difficulty && params.difficulty !== 'all') {
      queryParams.set('difficulty', params.difficulty);
    }
    if (params.category && params.category !== 'all') {
      queryParams.set('category', params.category);
    }
    if (params.type && params.type !== 'all') {
      queryParams.set('type', params.type);
    }
    const qs = queryParams.toString();
    const url = `${API_BASE}/problems${qs ? '?' + qs : ''}`;
    
    const response = await fetch(url);
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error?.message || `Failed to fetch problems: ${response.statusText}`);
    }
    const result = await response.json();
    return result.data || [];
  }

  async function fetchProblemDetail(id) {
    const url = `${API_BASE}/problems/${encodeURIComponent(id)}`;
    const response = await fetch(url);
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Problem '${id}' not found`);
      }
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error?.message || `Failed to fetch problem: ${response.statusText}`);
    }
    const result = await response.json();
    return result.data;
  }

  async function submitCode(problemId, code) {
    const url = `${API_BASE}/problems/${encodeURIComponent(problemId)}/run`;
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code })
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error?.message || `Request failed: ${response.statusText}`);
    }
    const result = await response.json();
    return result.data;
  }

  async function fetchHints(problemId) {
    const url = `${API_BASE}/problems/${encodeURIComponent(problemId)}/hints`;
    const response = await fetch(url);
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error?.message || `Failed to fetch hints: ${response.statusText}`);
    }
    const result = await response.json();
    return result.data?.hints || [];
  }

  // ---- Rendering Functions ----
  function renderProblemGrid() {
    const { filteredProblems } = state;
    
    if (filteredProblems.length === 0) {
      elements.problemGrid.style.display = 'none';
      elements.emptyState.style.display = 'block';
      return;
    }
    
    elements.problemGrid.style.display = 'grid';
    elements.emptyState.style.display = 'none';
    
    elements.problemGrid.innerHTML = filteredProblems.map(problem => {
      const difficultyClass = getDifficultyBadgeClass(problem.difficulty);
      const tagsHtml = (problem.tags || []).slice(0, 3).map(tag => 
        `<span class="tag">${escapeHtml(tag)}</span>`
      ).join('');
      const moreTags = (problem.tags || []).length > 3 ? 
        `<span class="tag">+${problem.tags.length - 3}</span>` : '';
      
      return `
        <article class="problem-card" data-problem-id="${escapeHtml(problem.id)}" onclick="window.app.navigateToProblem('${escapeHtml(problem.id)}')">
          <div class="problem-card-header">
            <h3 class="problem-card-title">${escapeHtml(problem.title)}</h3>
            <span class="badge ${difficultyClass}">${escapeHtml(problem.difficulty)}</span>
          </div>
          <div class="problem-card-meta">
            <span class="category-badge">${escapeHtml(problem.category)}</span>
            <span class="problem-card-type">${escapeHtml(problem.type || 'function')}</span>
          </div>
          ${tagsHtml || moreTags ? `<div class="problem-card-tags">${tagsHtml}${moreTags}</div>` : ''}
        </article>
      `;
    }).join('');
  }

  function renderFilters() {
    // Difficulty filter
    if (elements.filterDifficulty) {
      elements.filterDifficulty.innerHTML = state.difficulties.map(d => {
        const label = d === 'all' ? 'All Difficulties' : capitalize(d);
        const selected = state.filters.difficulty === d ? 'selected' : '';
        return `<option value="${d}" ${selected}>${label}</option>`;
      }).join('');
    }
    
    // Category filter
    if (elements.filterCategory) {
      const allCategories = ['all', ...state.categories];
      elements.filterCategory.innerHTML = allCategories.map(c => {
        const label = c === 'all' ? 'All Categories' : capitalize(c);
        const selected = state.filters.category === c ? 'selected' : '';
        return `<option value="${c}" ${selected}>${label}</option>`;
      }).join('');
    }
  }

  function renderProblemDetail() {
    const problem = state.currentProblem;
    if (!problem) return;
    
    // Update page title
    if (elements.pageTitle) {
      elements.pageTitle.textContent = problem.title;
    }
    document.title = `${problem.title} - Coding Challenge`;
    
    // Render examples
    const examplesHtml = (problem.examples || []).map((ex, i) => `
      <div class="example-block">
        <div class="example-header">Example ${i + 1}</div>
        <div class="example-io">
          <div class="example-io-item">
            <div class="example-io-label">Input</div>
            <div class="example-io-value">${formatCodeBlock(ex.input)}</div>
          </div>
          <div class="example-io-item">
            <div class="example-io-label">Output</div>
            <div class="example-io-value">${formatCodeBlock(ex.output)}</div>
          </div>
        </div>
        ${ex.explanation ? `
          <div class="example-explanation">${escapeHtml(ex.explanation)}</div>
        ` : ''}
      </div>
    `).join('');
    
    // Render constraints
    const constraintsHtml = (problem.constraints || []).length ? `
      <ul class="constraints-list">
        ${problem.constraints.map(c => `<li>${escapeHtml(c)}</li>`).join('')}
      </ul>
    ` : '';
    
    // Build the problem info panel
    if (elements.problemInfo) {
      elements.problemInfo.innerHTML = `
        <div class="problem-info-header">
          <h1>${escapeHtml(problem.title)}</h1>
          <span class="badge ${getDifficultyBadgeClass(problem.difficulty)}">${escapeHtml(problem.difficulty)}</span>
          <span class="category-badge">${escapeHtml(problem.category)}</span>
        </div>
        
        ${problem.type === 'function' && problem.function_name ? `
          <div class="function-signature">
            <span class="function-signature-label">Function:</span>
            <code>${escapeHtml(problem.function_name)}(${(problem.parameters || []).map(p => escapeHtml(p.name) + ' ' + escapeHtml(p.type)).join(', ')})</code>
            ${problem.return_type ? `<span class="function-signature-return">→ ${escapeHtml(problem.return_type)}</span>` : ''}
          </div>
        ` : ''}
        
        <div class="section">
          <div class="section-title">Description</div>
          <div class="problem-description">${formatDescription(problem.description)}</div>
        </div>
        
        ${examplesHtml ? `
          <div class="section">
            <div class="section-title">Examples</div>
            ${examplesHtml}
          </div>
        ` : ''}
        
        ${constraintsHtml ? `
          <div class="section">
            <div class="section-title">Constraints</div>
            ${constraintsHtml}
          </div>
        ` : ''}
        
        ${problem.time_complexity_hint || problem.space_complexity_hint ? `
          <div class="section">
            <div class="section-title">Complexity</div>
            ${problem.time_complexity_hint ? `<p style="font-size:0.85rem;margin-bottom:4px;"><strong>Time:</strong> ${escapeHtml(problem.time_complexity_hint)}</p>` : ''}
            ${problem.space_complexity_hint ? `<p style="font-size:0.85rem;"><strong>Space:</strong> ${escapeHtml(problem.space_complexity_hint)}</p>` : ''}
          </div>
        ` : ''}
        
        ${renderHintSection()}
      `;
    }
    
    // Initialize editor with template
    if (state.editor && problem.template) {
      state.editor.setCode(problem.template);
    }
  }

  function renderHintSection() {
    const totalHints = state.hints.length;
    if (totalHints === 0) return '';
    
    const revealedHints = state.hints.slice(0, state.revealedHintCount);
    const hintsHtml = revealedHints.map((hint, i) => `
      <div class="hint-item" data-hint-index="${i}">
        <div class="hint-item-title">
          <span class="hint-level-badge">L${hint.level}</span>
          ${escapeHtml(hint.title)}
        </div>
        <div class="hint-item-content">${formatHintContent(hint.content)}</div>
      </div>
    `).join('');
    
    const canRevealMore = state.revealedHintCount < totalHints;
    const nextHint = canRevealMore ? state.hints[state.revealedHintCount] : null;
    
    return `
      <div class="hint-section" id="hint-section">
        <div class="section-title">Hints</div>
        <div class="hint-counter">
          <span class="hint-counter-current">${state.revealedHintCount}</span>
          <span class="hint-counter-separator">/</span>
          <span class="hint-counter-total">${totalHints}</span>
          ${nextHint ? `<span class="hint-counter-next">— Next: ${escapeHtml(nextHint.title)}</span>` : ''}
        </div>
        ${hintsHtml}
        ${canRevealMore ? `
          <button class="btn btn--secondary btn--sm hint-reveal-btn" id="hint-reveal-btn" onclick="window.app.revealNextHint()">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
              <line x1="12" y1="17" x2="12.01" y2="17"/>
            </svg>
            Show Next Hint
          </button>
        ` : `
          <div class="hint-all-revealed">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            All hints revealed
          </div>
        `}
      </div>
    `;
  }

  // ---- Filter Logic ----
  function applyFilters() {
    let filtered = [...state.problems];
    
    if (state.filters.difficulty !== 'all') {
      filtered = filtered.filter(p => p.difficulty === state.filters.difficulty);
    }
    
    if (state.filters.category !== 'all') {
      filtered = filtered.filter(p => p.category === state.filters.category);
    }
    
    if (state.filters.search.trim()) {
      const search = state.filters.search.toLowerCase().trim();
      filtered = filtered.filter(p => 
        p.title.toLowerCase().includes(search) ||
        (p.tags || []).some(t => t.toLowerCase().includes(search)) ||
        p.category.toLowerCase().includes(search) ||
        p.id.toLowerCase().includes(search)
      );
    }
    
    state.filteredProblems = filtered;
    renderProblemGrid();
  }

  // ---- Navigation ----
  function navigateToProblem(problemId) {
    window.location.hash = `#problem/${problemId}`;
  }

  function navigateToList() {
    window.location.hash = '#problems';
  }

  async function handleHashChange() {
    const hash = window.location.hash || '#problems';
    
    if (hash.startsWith('#problem/')) {
      const problemId = hash.replace('#problem/', '');
      await showProblemDetail(problemId);
    } else {
      showProblemList();
    }
  }

  async function showProblemList() {
    if (elements.viewList) elements.viewList.style.display = 'block';
    if (elements.viewDetail) elements.viewDetail.style.display = 'none';
    
    // Reset scroll position
    window.scrollTo(0, 0);
    
    try {
      // Fetch with current filters
      state.problems = await fetchProblems({
        difficulty: state.filters.difficulty,
        category: state.filters.category
      });
      state.categories = [...new Set(state.problems.map(p => p.category))].sort();
      renderFilters();
      applyFilters();
      
      if (elements.loadingState) elements.loadingState.style.display = 'none';
    } catch (error) {
      console.error('Failed to load problems:', error);
      if (elements.loadingState) {
        elements.loadingState.innerHTML = `
          <div class="empty-state">
            <div class="empty-state-icon">⚠️</div>
            <div class="empty-state-title">Failed to load problems</div>
            <p>${escapeHtml(error.message)}</p>
            <button class="btn btn--secondary btn--sm" style="margin-top:12px;" onclick="window.location.reload()">Retry</button>
          </div>
        `;
      }
    }
  }

  async function showProblemDetail(problemId) {
    if (elements.viewList) elements.viewList.style.display = 'none';
    if (elements.viewDetail) elements.viewDetail.style.display = 'block';
    
    // Reset scroll position
    window.scrollTo(0, 0);
    
    // Show loading state
    if (elements.loadingStateDetail) elements.loadingStateDetail.style.display = 'flex';
    if (elements.problemDetail) elements.problemDetail.style.display = 'none';
    if (elements.errorState) elements.errorState.style.display = 'none';
    
    try {
      const [problem, hints] = await Promise.all([
        fetchProblemDetail(problemId),
        fetchHints(problemId).catch(() => []) // Don't fail if hints fail
      ]);
      
      state.currentProblem = problem;
      state.hints = hints;
      state.revealedHintCount = 0;
      state.testResults = null;
      
      if (elements.loadingStateDetail) elements.loadingStateDetail.style.display = 'none';
      if (elements.problemDetail) elements.problemDetail.style.display = 'grid';
      
      renderProblemDetail();
      
      // Initialize editor after DOM update
      setTimeout(() => {
        initEditor();
        // Clear previous results
        if (elements.testResults) {
          elements.testResults.innerHTML = `
            <div class="section" style="text-align:center;padding:40px;color:var(--color-text-muted);">
              <p>Submit your code to see test results</p>
            </div>
          `;
        }
      }, 50);
    } catch (error) {
      console.error('Failed to load problem:', error);
      if (elements.loadingStateDetail) elements.loadingStateDetail.style.display = 'none';
      if (elements.errorState) {
        elements.errorState.style.display = 'block';
        if (elements.errorTitle) elements.errorTitle.textContent = error.message.includes('not found') ? 'Problem Not Found' : 'Error Loading Problem';
        if (elements.errorMessage) elements.errorMessage.textContent = error.message;
      }
    }
  }

  // ---- Editor Initialization ----
  function initEditor() {
    const textarea = document.getElementById('code-editor');
    if (!textarea) return;
    
    // Use EditorModule if available
    if (typeof EditorModule !== 'undefined') {
      const template = state.currentProblem?.template || '';
      EditorModule.init(textarea, template);
      state.editor = EditorModule;
      return;
    }
    
    // Fallback: Check if CodeMirror is available
    if (typeof CodeMirror === 'undefined') {
      console.warn('CodeMirror not loaded, falling back to plain textarea');
      return;
    }
    
    // Destroy existing instance
    if (state.editor && state.editor.destroy) {
      state.editor.destroy();
    }
    
    state.editor = {
      cm: CodeMirror.fromTextArea(textarea, {
        mode: 'text/x-go',
        theme: 'dracula',
        lineNumbers: true,
        indentUnit: 4,
        tabSize: 4,
        indentWithTabs: false,
        lineWrapping: true,
        matchBrackets: true,
        autoCloseBrackets: true,
        styleActiveLine: true,
        viewportMargin: Infinity,
        placeholder: '// Write your Go code here...',
        extraKeys: {
          'Tab': function(cm) {
            if (cm.somethingSelected()) {
              cm.indentSelection('add');
            } else {
              cm.replaceSelection('    ');
            }
          },
          'Ctrl-Space': 'autocomplete'
        }
      }),
      getCode() { return this.cm.getValue(); },
      setCode(code) { this.cm.setValue(code); },
      resetToTemplate(template) { this.cm.setValue(template || ''); },
      refresh() { this.cm.refresh(); },
      focus() { this.cm.focus(); }
    };
    
    // Refresh after a short delay
    setTimeout(() => {
      state.editor.refresh();
    }, 100);
  }

  // ---- Code Submission ----
  async function handleSubmit() {
    if (!state.editor || !state.currentProblem || state.isSubmitting) return;
    
    const code = state.editor.getCode();
    if (!code.trim()) {
      showResults({ error: 'Please write some code before submitting.' });
      return;
    }
    
    state.isSubmitting = true;
    updateSubmitButton(true);
    
    // Show loading in results panel
    if (elements.testResults) {
      elements.testResults.innerHTML = `
        <div class="loading-state">
          <div class="spinner"></div>
          <span>Running tests...</span>
        </div>
      `;
    }
    
    try {
      const result = await submitCode(state.currentProblem.id, code);
      state.testResults = result;
      showResults(result);
    } catch (error) {
      showResults({ error: error.message });
    } finally {
      state.isSubmitting = false;
      updateSubmitButton(false);
    }
  }

  function updateSubmitButton(loading) {
    if (elements.btnSubmit) {
      elements.btnSubmit.disabled = loading;
      elements.btnSubmit.innerHTML = loading 
        ? '<div class="spinner" style="width:14px;height:14px;border-width:2px;"></div> Running...'
        : '▶ Submit';
    }
  }

  // ---- Show Results (delegates to ResultModule) ----
  function showResults(data) {
    if (typeof ResultModule !== 'undefined') {
      ResultModule.render(data, elements.testResults);
    } else {
      // Fallback if ResultModule is not loaded
      if (elements.testResults) {
        elements.testResults.innerHTML = `
          <div class="section">
            <div class="compilation-error">
              ResultModule not loaded.<br>
              <pre style="margin-top:8px;font-size:0.78rem;">${escapeHtml(JSON.stringify(data, null, 2))}</pre>
            </div>
          </div>
        `;
      }
    }
  }

  // ---- Reset Code ----
  function handleReset() {
    if (!state.editor || !state.currentProblem) return;
    
    // Show confirmation
    showConfirmModal(
      'Reset Code',
      'Are you sure you want to reset your code to the template? Your current changes will be lost.',
      () => {
        if (state.currentProblem.template) {
          state.editor.setCode(state.currentProblem.template);
        } else {
          state.editor.setCode('');
        }
        // Clear results
        if (elements.testResults) {
          elements.testResults.innerHTML = `
            <div class="section" style="text-align:center;padding:40px;color:var(--color-text-muted);">
              <p>Submit your code to see test results</p>
            </div>
          `;
        }
      }
    );
  }

  // ---- Hint Reveal ----
  function revealNextHint() {
    if (state.revealedHintCount >= state.hints.length) return;
    
    const nextHint = state.hints[state.revealedHintCount];
    
    showConfirmModal(
      'Reveal Hint?',
      `Are you sure you want to see the next hint?\n\nHint ${state.revealedHintCount + 1} of ${state.hints.length}\n"${nextHint.title}"\n\nRevealing hints reduces the learning challenge. Try solving the problem on your own first!`,
      () => {
        state.revealedHintCount++;
        renderProblemDetail();
        
        // Scroll to the newly revealed hint
        setTimeout(() => {
          const newHint = elements.problemInfo?.querySelector(`[data-hint-index="${state.revealedHintCount - 1}"]`);
          if (newHint) {
            newHint.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            newHint.classList.add('hint-item--highlight');
            setTimeout(() => newHint.classList.remove('hint-item--highlight'), 2000);
          }
        }, 100);
      }
    );
  }

  // ---- Confirm Modal ----
  function showConfirmModal(title, body, onConfirm) {
    if (!elements.confirmModal) {
      // Fallback to native confirm if modal not available
      if (confirm(title + '\n\n' + body)) {
        onConfirm();
      }
      return;
    }
    
    elements.modalTitle.textContent = title;
    elements.modalBody.textContent = body;
    elements.confirmModal.classList.add('active');
    
    // Remove old event listeners by cloning
    const newBtnConfirm = elements.modalConfirm.cloneNode(true);
    elements.modalConfirm.parentNode.replaceChild(newBtnConfirm, elements.modalConfirm);
    elements.modalConfirm = newBtnConfirm;
    
    elements.modalConfirm.addEventListener('click', () => {
      elements.confirmModal.classList.remove('active');
      onConfirm();
    });
    
    elements.modalCancel.addEventListener('click', () => {
      elements.confirmModal.classList.remove('active');
    });
    
    elements.confirmModal.addEventListener('click', (e) => {
      if (e.target === elements.confirmModal) {
        elements.confirmModal.classList.remove('active');
      }
    });
  }

  // ---- Sidebar Toggle (Mobile) ----
  function toggleSidebar() {
    elements.sidebar.classList.toggle('open');
    elements.sidebarOverlay.classList.toggle('active');
    document.body.classList.toggle('sidebar-open');
  }

  function closeSidebar() {
    elements.sidebar.classList.remove('open');
    elements.sidebarOverlay.classList.remove('active');
    document.body.classList.remove('sidebar-open');
  }

  // ---- Utility Functions ----
  function getDifficultyBadgeClass(difficulty) {
    const map = {
      'easy': 'badge--easy',
      'medium': 'badge--medium',
      'hard': 'badge--hard',
      'expert': 'badge--expert'
    };
    return map[difficulty] || 'badge--easy';
  }

  function capitalize(str) {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
  }

  function escapeHtml(str) {
    if (str === null || str === undefined) return '';
    const div = document.createElement('div');
    div.textContent = String(str);
    return div.innerHTML;
  }

  function formatCodeBlock(str) {
    if (!str) return '';
    // Preserve formatting in code-like values
    return escapeHtml(str).replace(/\n/g, '<br>');
  }

  function formatDescription(str) {
    if (!str) return '';
    // Basic markdown-like formatting
    let html = escapeHtml(str);
    // Bold
    html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    // Inline code
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');
    // Line breaks
    html = html.replace(/\n/g, '<br>');
    return html;
  }

  function formatHintContent(content) {
    if (!content) return '';
    let html = escapeHtml(content);
    // Inline code
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');
    // Bold
    html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    // Line breaks
    html = html.replace(/\n/g, '<br>');
    return html;
  }

  // ---- Health Check ----
  async function checkHealth() {
    try {
      const response = await fetch('/health');
      const result = await response.json();
      if (elements.healthStatus) {
        const isOk = result.data?.status === 'ok';
        elements.healthStatus.className = `badge ${isOk ? 'badge--easy' : 'badge--medium'}`;
        elements.healthStatus.innerHTML = `
          <svg width="8" height="8" viewBox="0 0 8 8" fill="currentColor"><circle cx="4" cy="4" r="4"/></svg>
          ${isOk ? 'Online' : 'Degraded'}
        `;
      }
    } catch (e) {
      if (elements.healthStatus) {
        elements.healthStatus.className = 'badge badge--hard';
        elements.healthStatus.innerHTML = `
          <svg width="8" height="8" viewBox="0 0 8 8" fill="currentColor"><circle cx="4" cy="4" r="4"/></svg>
          Offline
        `;
      }
    }
  }

  // ---- Event Listeners ----
  function bindEvents() {
    // Filter changes
    if (elements.filterDifficulty) {
      elements.filterDifficulty.addEventListener('change', (e) => {
        state.filters.difficulty = e.target.value;
        // Also update sidebar
        document.querySelectorAll('.sidebar-link[data-difficulty]').forEach(link => {
          link.classList.toggle('active', link.dataset.difficulty === e.target.value);
        });
        applyFilters();
      });
    }
    
    if (elements.filterCategory) {
      elements.filterCategory.addEventListener('change', (e) => {
        state.filters.category = e.target.value;
        applyFilters();
      });
    }
    
    if (elements.filterSearch) {
      // Debounce search
      let searchTimeout;
      elements.filterSearch.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
          state.filters.search = e.target.value;
          applyFilters();
        }, 300);
      });
    }
    
    // Sidebar difficulty links
    document.querySelectorAll('.sidebar-link[data-difficulty]').forEach(link => {
      link.addEventListener('click', (e) => {
        e.preventDefault();
        const difficulty = link.dataset.difficulty;
        state.filters.difficulty = difficulty;
        if (elements.filterDifficulty) {
          elements.filterDifficulty.value = difficulty;
        }
        // Update active state
        document.querySelectorAll('.sidebar-link[data-difficulty]').forEach(l => l.classList.remove('active'));
        link.classList.add('active');
        applyFilters();
        closeSidebar();
      });
    });
    
    // Submit button
    if (elements.btnSubmit) {
      elements.btnSubmit.addEventListener('click', handleSubmit);
    }
    
    // Reset button
    if (elements.btnReset) {
      elements.btnReset.addEventListener('click', handleReset);
    }

    // Shortcuts button
    var btnShortcuts = document.getElementById('btn-shortcuts');
    if (btnShortcuts) {
      btnShortcuts.addEventListener('click', toggleKeyboardShortcuts);
    }
    
    // Back link
    if (elements.backLink) {
      elements.backLink.addEventListener('click', navigateToList);
    }
    
    // Sidebar toggle
    if (elements.sidebarToggle) {
      elements.sidebarToggle.addEventListener('click', toggleSidebar);
    }
    if (elements.sidebarOverlay) {
      elements.sidebarOverlay.addEventListener('click', closeSidebar);
    }
    
    // Hash change for routing
    window.addEventListener('hashchange', handleHashChange);
    
    // Keyboard shortcut: Ctrl+Enter to submit
    document.addEventListener('keydown', (e) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        if (elements.viewDetail && elements.viewDetail.style.display !== 'none') {
          e.preventDefault();
          handleSubmit();
        }
      }
      // Keyboard shortcut: Ctrl+/ to show keyboard shortcuts help
      if ((e.ctrlKey || e.metaKey) && e.key === '/') {
        e.preventDefault();
        toggleKeyboardShortcuts();
      }
      // Keyboard shortcut: Escape to close modals/panels
      if (e.key === 'Escape') {
        closeKeyboardShortcuts();
        closeConfirmModal();
      }
    });
    
    // Close sidebar on window resize to desktop
    window.addEventListener('resize', () => {
      if (window.innerWidth > 768) {
        closeSidebar();
      }
    });
  }

  // ---- Keyboard Shortcuts Help Panel ----
  function toggleKeyboardShortcuts() {
    const existingPanel = document.getElementById('shortcuts-panel');
    if (existingPanel) {
      existingPanel.remove();
      return;
    }

    const panel = document.createElement('div');
    panel.id = 'shortcuts-panel';
    panel.className = 'shortcuts-panel';
    panel.innerHTML = `
      <div class="shortcuts-header">
        <h3>⌨️ Keyboard Shortcuts</h3>
        <button class="shortcuts-close" onclick="document.getElementById('shortcuts-panel').remove()">✕</button>
      </div>
      <div class="shortcuts-list">
        <div class="shortcut-item">
          <kbd>Ctrl</kbd> + <kbd>Enter</kbd>
          <span class="shortcut-desc">Submit code</span>
        </div>
        <div class="shortcut-item">
          <kbd>Ctrl</kbd> + <kbd>R</kbd>
          <span class="shortcut-desc">Reset to template</span>
        </div>
        <div class="shortcut-item">
          <kbd>Ctrl</kbd> + <kbd>/</kbd>
          <span class="shortcut-desc">Show shortcuts</span>
        </div>
        <div class="shortcut-item">
          <kbd>Esc</kbd>
          <span class="shortcut-desc">Close panel/modal</span>
        </div>
        <div class="shortcut-item">
          <kbd>Tab</kbd>
          <span class="shortcut-desc">Indent / autocomplete</span>
        </div>
        <div class="shortcut-item">
          <kbd>Ctrl</kbd> + <kbd>Space</kbd>
          <span class="shortcut-desc">Trigger autocomplete</span>
        </div>
      </div>
    `;
    document.body.appendChild(panel);
  }

  function closeKeyboardShortcuts() {
    const panel = document.getElementById('shortcuts-panel');
    if (panel) panel.remove();
  }

  function closeConfirmModal() {
    if (elements.confirmModal) {
      elements.confirmModal.classList.remove('active');
    }
  }

  // ---- Initialize Application ----
  async function init() {
    cacheElements();
    bindEvents();
    
    // Check health
    checkHealth();
    setInterval(checkHealth, 30000);
    
    // Handle initial route
    await handleHashChange();
  }
  
  // ---- Expose Public API ----
  window.app = {
    navigateToProblem,
    revealNextHint,
    toggleSidebar,
    closeSidebar
  };

  // ---- Boot ----
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
