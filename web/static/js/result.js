/**
 * Test Result Display Module
 * Coding Challenge Website
 * 
 * Handles rendering of test results including pass/fail status,
 * expected vs actual output, compilation errors, and progress bar.
 */

const ResultModule = (function() {
  'use strict';

  /**
   * Render test results into the container
   * @param {Object} data - The result data from API or error
   * @param {boolean} data.success - Whether all tests passed
   * @param {string|null} data.compilation_error - Compilation error message
   * @param {Array} data.test_results - Array of test case results
   * @param {number} data.passed_count - Number of passed tests
   * @param {number} data.total_count - Total number of tests
   * @param {number} data.execution_time_ms - Execution time in ms
   * @param {string} [data.error] - General error message
   * @param {HTMLElement} [container] - Optional container element
   */
  function render(data, container) {
    container = container || document.getElementById('test-results');
    if (!container) return;

    // Handle general error (not from API response)
    if (data.error && !data.test_results && !data.compilation_error) {
      container.innerHTML = `
        <div class="section">
          <div class="compilation-error">
            <strong>Error:</strong> ${escapeHtml(data.error)}
          </div>
        </div>
      `;
      return;
    }

    const passedCount = data.passed_count || 0;
    const totalCount = data.total_count || 0;
    const allPassed = data.success === true;
    const hasCompilationError = !!data.compilation_error;
    const testResults = data.test_results || [];
    const executionTime = data.execution_time_ms || 0;

    // Build summary section
    const summaryHtml = buildSummary(allPassed, hasCompilationError, passedCount, totalCount, executionTime);

    // Build compilation error section
    const compilationHtml = hasCompilationError ? buildCompilationError(data.compilation_error) : '';

    // Build test cases section
    const testCasesHtml = testResults.length > 0 ? buildTestCases(testResults) : '';

    // Build empty state for compilation error with no tests
    const emptyTestsHtml = hasCompilationError && testResults.length === 0 ? `
      <div class="section" style="margin-top:12px;">
        <p style="font-size:0.85rem;color:var(--color-text-muted);">
          No tests were executed due to compilation error.
        </p>
      </div>
    ` : '';

    container.innerHTML = `
      <div class="results-panel">
        ${summaryHtml}
        ${compilationHtml}
        ${testCasesHtml}
        ${emptyTestsHtml}
      </div>
    `;

    // Trigger confetti if all tests passed
    if (allPassed && totalCount > 0) {
      triggerConfetti();
    }

    // Bind toggle events for test case details
    bindToggleEvents();
  }

  /**
   * Build the summary section with status, progress bar, and stats
   */
  function buildSummary(allPassed, hasCompilationError, passedCount, totalCount, executionTime) {
    const statusText = allPassed ? 'All Tests Passed!' : 'Some Tests Failed';
    const statusClass = allPassed ? 'results-status--passed' : 'results-status--failed';
    const statusIcon = allPassed ? '✓' : '✗';
    const progressClass = allPassed ? 'progress-bar-fill--passed' : 'progress-bar-fill--failed';
    const progressPercent = totalCount > 0 ? Math.round((passedCount / totalCount) * 100) : 0;

    return `
      <div class="section">
        <div class="results-summary">
          <div class="results-status ${statusClass}">
            <span class="results-status-icon">${statusIcon}</span>
            <span>${statusText}</span>
          </div>
          <div class="progress-bar-container">
            <div class="progress-bar-track">
              <div class="progress-bar-fill ${progressClass}" style="width: ${progressPercent}%"></div>
            </div>
            <div class="progress-bar-text">
              ${passedCount} / ${totalCount} tests passed (${progressPercent}%)
            </div>
          </div>
          ${executionTime > 0 ? `
            <div class="execution-time">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <polyline points="12 6 12 12 16 14"/>
              </svg>
              ${executionTime}ms
            </div>
          ` : ''}
        </div>
      </div>
    `;
  }

  /**
   * Build compilation error display with syntax highlighting
   */
  function buildCompilationError(error) {
    return `
      <div class="section">
        <div class="section-title">Compilation Error</div>
        <div class="compilation-error syntax-highlighted">${highlightGoError(error)}</div>
      </div>
    `;
  }

  /**
   * Highlight Go compilation errors with basic syntax highlighting
   */
  function highlightGoError(error) {
    let html = escapeHtml(error);
    // Highlight error lines (e.g., "main.go:10:5: error message")
    html = html.replace(/(main\.go):(\d+):(\d+):/g, '<span style="color:#f87171;font-weight:600;">$1:$2:$3:</span>');
    // Highlight the word "error"
    html = html.replace(/\b(error)\b/g, '<span style="color:#f87171;font-weight:600;">$1</span>');
    // Highlight Go keywords in error messages
    html = html.replace(/\b(undefined|cannot|expected|found|declared|nil)\b/g, '<span style="color:#fbbf24;">$1</span>');
    return html;
  }

  /**
   * Build individual test case result items
   */
  function buildTestCases(testResults) {
    const itemsHtml = testResults.map((test, index) => {
      const statusClass = test.passed ? 'test-case--passed' : 'test-case--failed';
      const statusIcon = test.passed 
        ? '<svg class="test-case-status-icon" viewBox="0 0 24 24" fill="none" stroke="#16a34a" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>'
        : '<svg class="test-case-status-icon" viewBox="0 0 24 24" fill="none" stroke="#dc2626" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>';
      
      const name = test.name || `Test Case ${index + 1}`;
      const errorHtml = test.error ? `<div class="test-case-error">${escapeHtml(test.error)}</div>` : '';
      
      // Format expected/actual values - handle objects/arrays properly
      const expectedDisplay = formatValue(test.expected);
      const actualDisplay = formatValue(test.actual);
      const execTime = test.execution_time_ms ? `<span style="margin-left:8px;color:var(--color-text-muted);">${test.execution_time_ms}ms</span>` : '';

      return `
        <div class="test-case ${statusClass}" data-test-index="${index}">
          <div class="test-case-header" onclick="ResultModule.toggleTestCase(${index})">
            ${statusIcon}
            <span class="test-case-name">${escapeHtml(name)}</span>
            ${execTime}
            <span class="test-case-toggle">▼</span>
          </div>
          <div class="test-case-body" id="test-case-body-${index}">
            <div class="test-case-row">
              <span class="test-case-label">Expected</span>
              <span class="test-case-value">${expectedDisplay}</span>
            </div>
            <div class="test-case-row">
              <span class="test-case-label">Actual</span>
              <span class="test-case-value">${actualDisplay}</span>
            </div>
            ${errorHtml ? `
              <div class="test-case-row">
                <span class="test-case-label">Error</span>
                ${errorHtml}
              </div>
            ` : ''}
          </div>
        </div>
      `;
    }).join('');

    return `
      <div class="section">
        <div class="section-title">Test Cases</div>
        ${itemsHtml}
      </div>
    `;
  }

  /**
   * Format a value for display - handles objects, arrays, null, etc.
   * @param {*} value
   * @returns {string} Formatted HTML
   */
  function formatValue(value) {
    if (value === null || value === undefined) {
      return '<em style="color:var(--color-text-muted);">(no output)</em>';
    }
    if (typeof value === 'object') {
      // Pretty-print JSON for objects/arrays
      try {
        const json = JSON.stringify(value, null, 2);
        return `<pre style="margin:0;padding:4px 8px;font-size:0.78rem;background:var(--color-bg);border-radius:4px;white-space:pre-wrap;font-family:var(--font-mono);">${escapeHtml(json)}</pre>`;
      } catch (e) {
        return escapeHtml(String(value));
      }
    }
    // For simple string values, preserve formatting
    const str = String(value);
    if (str.includes('\n')) {
      return `<pre style="margin:0;padding:4px 8px;font-size:0.78rem;background:var(--color-bg);border-radius:4px;white-space:pre-wrap;font-family:var(--font-mono);">${escapeHtml(str)}</pre>`;
    }
    return escapeHtml(str);
  }

  /**
   * Toggle test case body visibility
   * @param {number} index - Test case index
   */
  function toggleTestCase(index) {
    const body = document.getElementById(`test-case-body-${index}`);
    const header = body ? body.previousElementSibling : null;
    
    if (body) {
      body.classList.toggle('expanded');
      const toggle = header ? header.querySelector('.test-case-toggle') : null;
      if (toggle) {
        toggle.textContent = body.classList.contains('expanded') ? '▲' : '▼';
      }
    }
  }

  /**
   * Bind toggle events for test case headers
   */
  function bindToggleEvents() {
    // Events are bound via onclick in HTML for simplicity
    // This is kept for future use if we need delegated events
  }

  /**
   * Show loading state in results panel
   */
  function showLoading() {
    const container = document.getElementById('test-results');
    if (container) {
      container.innerHTML = `
        <div class="loading-state">
          <div class="spinner"></div>
          <span>Running tests...</span>
        </div>
      `;
    }
  }

  /**
   * Show error state in results panel
   * @param {string} message - Error message
   */
  function showError(message) {
    const container = document.getElementById('test-results');
    if (container) {
      container.innerHTML = `
        <div class="section">
          <div class="compilation-error">
            <strong>Error:</strong> ${escapeHtml(message)}
          </div>
        </div>
      `;
    }
  }

  /**
   * Clear results panel
   */
  function clear() {
    const container = document.getElementById('test-results');
    if (container) {
      container.innerHTML = `
        <div class="section" style="text-align:center;padding:40px;color:var(--color-text-muted);">
          <p>Submit your code to see test results</p>
        </div>
      `;
    }
  }

  /**
   * Escape HTML entities
   */
  function escapeHtml(str) {
    if (str === null || str === undefined) return '';
    const div = document.createElement('div');
    div.textContent = String(str);
    return div.innerHTML;
  }

  /**
   * Trigger confetti animation when all tests pass
   */
  function triggerConfetti() {
    const colors = ['#16a34a', '#2563eb', '#fbbf24', '#dc2626', '#7c3aed', '#ec4899'];
    const container = document.body;
    const confettiCount = 80;

    for (let i = 0; i < confettiCount; i++) {
      const confetti = document.createElement('div');
      confetti.className = 'confetti-particle';
      confetti.style.cssText = `
        position: fixed;
        width: ${Math.random() * 10 + 5}px;
        height: ${Math.random() * 6 + 4}px;
        background: ${colors[Math.floor(Math.random() * colors.length)]};
        left: ${Math.random() * 100}vw;
        top: -20px;
        border-radius: ${Math.random() > 0.5 ? '50%' : '2px'};
        z-index: 9999;
        pointer-events: none;
        animation: confettiFall ${Math.random() * 2 + 2}s ease-out forwards;
        animation-delay: ${Math.random() * 0.5}s;
      `;
      container.appendChild(confetti);

      // Remove after animation
      setTimeout(() => {
        confetti.remove();
      }, 4000);
    }
  }

  // Public API
  return {
    render,
    showLoading,
    showError,
    clear,
    toggleTestCase,
    formatValue,
    triggerConfetti
  };
})();
