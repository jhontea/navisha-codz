/**
 * CodeMirror Editor Integration
 * Coding Challenge Website
 * 
 * Handles CodeMirror initialization, configuration, and operations
 * for the Go code editor.
 */

const EditorModule = (function() {
  'use strict';

  let editorInstance = null;
  const defaultTemplate = `func solution() {\n    // Your code here\n}`;

  /**
   * Initialize CodeMirror editor
   * @param {string|HTMLTextAreaElement} target - Selector or textarea element
   * @param {string} template - Initial code template
   * @returns {CodeMirror.Editor} The editor instance
   */
  function init(target, template) {
    const textarea = typeof target === 'string' ? document.querySelector(target) : target;
    if (!textarea) {
      console.error('Editor target not found:', target);
      return null;
    }

    // Destroy existing instance if any
    if (editorInstance) {
      try {
        editorInstance.toTextArea();
      } catch (e) { /* ignore */ }
      editorInstance = null;
    }

    // Check if CodeMirror is loaded
    if (typeof CodeMirror === 'undefined') {
      console.warn('CodeMirror is not loaded. Ensure CDN scripts are included.');
      return null;
    }

    // Ensure Go mode is loaded
    if (!CodeMirror.modes['text/x-go']) {
      console.warn('Go mode not loaded. Using plain text mode.');
    }

    editorInstance = CodeMirror.fromTextArea(textarea, {
      // Core settings
      mode: 'text/x-go',
      theme: 'dracula',
      lineNumbers: true,
      
      // Indentation
      indentUnit: 4,
      tabSize: 4,
      indentWithTabs: false,
      
      // Visual
      lineWrapping: true,
      styleActiveLine: true,
      showTrailingSpace: true,
      
      // Functionality
      matchBrackets: true,
      autoCloseBrackets: true,
      
      // Performance
      viewportMargin: Infinity,
      
      // Placeholder when empty
      placeholder: '// Write your Go code here...',
      
      // Extra key bindings
      extraKeys: {
        'Tab': function(cm) {
          if (cm.somethingSelected()) {
            cm.indentSelection('add');
          } else {
            cm.replaceSelection('    ', 'end');
          }
        },
        'Shift-Tab': function(cm) {
          if (cm.somethingSelected()) {
            cm.indentSelection('subtract');
          }
        },
        'Ctrl-Space': 'autocomplete',
        'Ctrl-/': 'toggleComment',
        'Cmd-/': 'toggleComment',
        'Enter': function(cm) {
          // Auto-indent on Enter
          const cursor = cm.getCursor();
          const line = cm.getLine(cursor.line);
          const match = line.match(/^(\s*)/);
          const indent = match ? match[1] : '';
          
          // Add extra indent after { or (
          const lastChar = line.trim().slice(-1);
          const extraIndent = (lastChar === '{' || lastChar === '(' || lastChar === ':') ? '    ' : '';
          
          cm.replaceSelection('\n' + indent + extraIndent, 'end');
        }
      }
    });

    // Set initial value
    if (template) {
      editorInstance.setValue(template);
    } else {
      editorInstance.setValue(defaultTemplate);
    }

    // Refresh after short delay for proper rendering in hidden containers
    setTimeout(() => {
      if (editorInstance) {
        editorInstance.refresh();
      }
    }, 100);

    // Handle editor resize
    const container = editorInstance.getWrapperElement();
    container.style.minHeight = '400px';

    return editorInstance;
  }

  /**
   * Get the current editor instance
   * @returns {CodeMirror.Editor|null}
   */
  function getInstance() {
    return editorInstance;
  }

  /**
   * Get the current code content
   * @returns {string}
   */
  function getCode() {
    if (editorInstance) {
      return editorInstance.getValue();
    }
    return '';
  }

  /**
   * Set code content
   * @param {string} code
   */
  function setCode(code) {
    if (editorInstance) {
      editorInstance.setValue(code || '');
    }
  }

  /**
   * Reset editor to template
   * @param {string} template
   */
  function resetToTemplate(template) {
    if (editorInstance) {
      editorInstance.setValue(template || defaultTemplate);
    }
  }

  /**
   * Refresh the editor (use after container resize or visibility change)
   */
  function refresh() {
    if (editorInstance) {
      editorInstance.refresh();
    }
  }

  /**
   * Focus the editor
   */
  function focus() {
    if (editorInstance) {
      editorInstance.focus();
    }
  }

  /**
   * Get selected text
   * @returns {string}
   */
  function getSelection() {
    if (editorInstance) {
      return editorInstance.getSelection();
    }
    return '';
  }

  /**
   * Replace selection with text
   * @param {string} text
   */
  function replaceSelection(text) {
    if (editorInstance) {
      editorInstance.replaceSelection(text);
    }
  }

  /**
   * Execute a command on the editor
   * @param {string} command
   */
  function execCommand(command) {
    if (editorInstance) {
      editorInstance.execCommand(command);
    }
  }

  /**
   * Destroy the editor instance
   */
  function destroy() {
    if (editorInstance) {
      try {
        editorInstance.toTextArea();
      } catch (e) { /* ignore */ }
      editorInstance = null;
    }
  }

  /**
   * Check if editor is ready
   * @returns {boolean}
   */
  function isReady() {
    return editorInstance !== null;
  }

  // Public API
  return {
    init,
    getInstance,
    getCode,
    setCode,
    resetToTemplate,
    refresh,
    focus,
    getSelection,
    replaceSelection,
    execCommand,
    destroy,
    isReady
  };
})();
