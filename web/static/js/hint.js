/**
 * Hint Reveal UI Module
 * Coding Challenge Website
 * 
 * Handles progressive hint reveal with confirmation dialog,
 * hint counter display, and hint content rendering.
 * 
 * Note: This module works in conjunction with app.js which handles
 * the actual hint rendering and confirmation flow. This module
 * provides utility functions for hint formatting.
 */

const HintModule = (function() {
  'use strict';

  let hints = [];
  let revealedCount = 0;
  let onRevealCallback = null;

  /**
   * Initialize hint module
   * @param {Array} hintList - Array of hint objects from API
   * @param {Function} onReveal - Callback when a hint is revealed
   */
  function init(hintList, onReveal) {
    hints = hintList || [];
    revealedCount = 0;
    onRevealCallback = onReveal;
    render();
  }

  /**
   * Render the hint section
   * @returns {string} HTML string
   */
  function render() {
    const totalHints = hints.length;
    if (totalHints === 0) return '';

    const revealedHints = hints.slice(0, revealedCount);
    const hintsHtml = revealedHints.map((hint, i) => `
      <div class="hint-item" data-hint-index="${i}">
        <div class="hint-item-title">
          <span class="hint-level-badge">L${hint.level}</span>
          ${escapeHtml(hint.title)}
        </div>
        <div class="hint-item-content">${formatHintContent(hint.content)}</div>
      </div>
    `).join('');

    const canRevealMore = revealedCount < totalHints;
    const nextHint = canRevealMore ? hints[revealedCount] : null;

    return `
      <div class="hint-section" id="hint-section">
        <div class="section-title">Hints</div>
        <div class="hint-counter">
          <span class="hint-counter-current">${revealedCount}</span>
          <span class="hint-counter-separator">/</span>
          <span class="hint-counter-total">${totalHints}</span>
          ${nextHint ? `<span class="hint-counter-next">— Next: ${escapeHtml(nextHint.title)}</span>` : ''}
        </div>
        ${hintsHtml}
        ${canRevealMore ? `
          <button class="btn btn--secondary btn--sm hint-reveal-btn" id="hint-reveal-btn" onclick="HintModule.requestReveal()">
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

  /**
   * Request to reveal the next hint (shows confirmation)
   */
  function requestReveal() {
    if (revealedCount >= hints.length) return;

    const nextHint = hints[revealedCount];
    const hintNumber = revealedCount + 1;
    const totalHints = hints.length;

    // Show confirmation dialog
    const confirmed = confirm(
      `Are you sure you want to see the next hint?\n\n` +
      `Hint ${hintNumber} of ${totalHints}\n` +
      `"${nextHint.title}"\n\n` +
      `Revealing hints reduces the learning challenge. ` +
      `Try solving the problem on your own first!`
    );

    if (confirmed) {
      revealNext();
    }
  }

  /**
   * Reveal the next hint without confirmation
   */
  function revealNext() {
    if (revealedCount >= hints.length) return;
    
    revealedCount++;
    render();
    
    // Scroll to the newly revealed hint
    setTimeout(() => {
      const section = document.getElementById('hint-section');
      if (section) {
        const hintItems = section.querySelectorAll('.hint-item');
        if (hintItems.length > 0) {
          const latestHint = hintItems[hintItems.length - 1];
          latestHint.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
          latestHint.classList.add('hint-item--highlight');
          setTimeout(() => latestHint.classList.remove('hint-item--highlight'), 2000);
        }
      }
    }, 50);

    // Trigger callback
    if (onRevealCallback) {
      onRevealCallback(revealedCount, hints[revealedCount - 1]);
    }
  }

  /**
   * Reveal all hints at once (with confirmation)
   */
  function revealAll() {
    if (revealedCount >= hints.length) return;

    const confirmed = confirm(
      `Are you sure you want to reveal ALL remaining hints?\n\n` +
      `This will reveal ${hints.length - revealedCount} more hint(s).\n` +
      `This significantly reduces the learning experience.`
    );

    if (confirmed) {
      revealedCount = hints.length;
      render();
      
      if (onRevealCallback) {
        onRevealCallback(revealedCount, null);
      }
    }
  }

  /**
   * Reset all hints (hide all)
   */
  function reset() {
    revealedCount = 0;
    render();
  }

  /**
   * Get current hint state
   * @returns {Object} { revealedCount, totalHints, canRevealMore }
   */
  function getState() {
    return {
      revealedCount,
      totalHints: hints.length,
      canRevealMore: revealedCount < hints.length,
      nextHint: revealedCount < hints.length ? hints[revealedCount] : null
    };
  }

  /**
   * Format hint content (handle basic markdown-like formatting)
   * @param {string} content
   * @returns {string} Formatted HTML
   */
  function formatHintContent(content) {
    if (!content) return '';
    
    let html = escapeHtml(content);
    
    // Handle inline code
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');
    
    // Handle bold
    html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    
    // Handle italic
    html = html.replace(/\*([^*]+)\*/g, '<em>$1</em>');
    
    // Handle line breaks
    html = html.replace(/\n/g, '<br>');
    
    return html;
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

  // Public API
  return {
    init,
    render,
    requestReveal,
    revealNext,
    revealAll,
    reset,
    getState,
    formatHintContent
  };
})();
