import React, { useRef, useEffect, useCallback, useMemo, memo, useState } from "react";
import Editor, { OnMount, DiffEditor } from "@monaco-editor/react";
import * as monaco from "monaco-editor";
import type { editor } from "monaco-editor";
import { Play, Copy, RotateCcw, Palette, AlignLeft, Minimize2, GitCompare } from "lucide-react";
import { useDebounce } from "../hooks/useDebounce";

// ── Constants ───────────────────────────────────────────────────────────────

type EditorTheme = "vs-dark" | "light" | "monokai" | "dracula" | "github-light";

interface ThemeOption {
  id: EditorTheme;
  label: string;
  monacoTheme: string;
}

const THEME_OPTIONS: ThemeOption[] = [
  { id: "vs-dark", label: "VS Code Dark", monacoTheme: "vs-dark" },
  { id: "light", label: "VS Code Light", monacoTheme: "light" },
  { id: "monokai", label: "Monokai", monacoTheme: "monokai" },
  { id: "dracula", label: "Dracula", monacoTheme: "dracula" },
  { id: "github-light", label: "GitHub Light", monacoTheme: "github-light" },
];

const FONT_SIZES = [12, 13, 14, 15, 16, 18, 20, 22, 24];

const GO_SNIPPETS: Omit<monaco.languages.CompletionItem, 'range'>[] = [
  {
    label: "for-range",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["for ${1:i}, ${2:v} := range ${3:slice} {", "\t$0", "}"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "For range loop over slice/map",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "if-err",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["if err != nil {", "\t$0", "}"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Error handling pattern",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "func-main",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["func main() {", "\t$0", "}"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Main function",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "switch",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["switch ${1:expr} {", "case ${2:val}:", "\t$0", "}"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Switch statement",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "if",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["if ${1:condition} {", "\t$0", "}"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "If statement",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "for",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["for ${1:i} := 0; $1 < ${2:n}; $1++ {", "\t$0", "}"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "For loop",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "func",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["func ${1:name}(${2:args}) ${3:returnType} {", "\t$0", "}"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Function declaration",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "struct",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["type ${1:Name} struct {", "\t$0", "}"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Struct type",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "fmt.Printf",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: 'fmt.Printf("${1:%v}\\\\n", ${2:val})',
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Printf debugging",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "defer",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: ["defer ${1:func}()"].join("\n"),
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Defer function call",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "slice-make",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: "${1:s} := make([]${2:type}, ${3:len}, ${4:cap})",
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Make a slice",
    detail: "Go snippet",
    sortText: "0",
  },
  {
    label: "map-make",
    kind: monaco.languages.CompletionItemKind.Snippet,
    insertText: "${1:m} := make(map[${2:keyType}]${3:valType})",
    insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
    documentation: "Make a map",
    detail: "Go snippet",
    sortText: "0",
  },
];

interface CodeEditorProps {
  value: string;
  onChange: (value: string) => void;
  language?: string;
  theme?: EditorTheme;
  readOnly?: boolean;
  template?: string;
  onSubmit?: () => void;
  previousSubmission?: string | null;
  errorLines?: number[];
  height?: string;
}

// ── Custom Monaco Theme Definition ──────────────────────────────────────────

function defineCustomThemes(monacoInstance: typeof monaco) {
  // Monokai
  monacoInstance.editor.defineTheme("monokai", {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "comment", foreground: "88846f", fontStyle: "italic" },
      { token: "keyword", foreground: "f92672" },
      { token: "string", foreground: "e6db74" },
      { token: "number", foreground: "ae81ff" },
      { token: "type", foreground: "66d9ef" },
      { token: "function", foreground: "a6e22e" },
      { token: "variable", foreground: "f8f8f2" },
    ],
    colors: {
      "editor.background": "#272822",
      "editor.foreground": "#f8f8f2",
      "editor.lineHighlightBackground": "#3e3d32",
      "editorCursor.foreground": "#f8f8f2",
      "editor.selectionBackground": "#49483e",
      "editor.inactiveSelectionBackground": "#3e3d32",
    },
  });

  // Dracula
  monacoInstance.editor.defineTheme("dracula", {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "comment", foreground: "6272a4", fontStyle: "italic" },
      { token: "keyword", foreground: "ff79c6" },
      { token: "string", foreground: "f1fa8c" },
      { token: "number", foreground: "bd93f9" },
      { token: "type", foreground: "8be9fd" },
      { token: "function", foreground: "50fa7b" },
    ],
    colors: {
      "editor.background": "#282a36",
      "editor.foreground": "#f8f8f2",
      "editor.lineHighlightBackground": "#44475a",
      "editorCursor.foreground": "#f8f8f2",
      "editor.selectionBackground": "#44475a",
    },
  });

  // GitHub Light
  monacoInstance.editor.defineTheme("github-light", {
    base: "vs",
    inherit: true,
    rules: [
      { token: "comment", foreground: "6e7781", fontStyle: "italic" },
      { token: "keyword", foreground: "cf222e" },
      { token: "string", foreground: "0a3069" },
      { token: "number", foreground: "0550ae" },
      { token: "type", foreground: "6639ba" },
      { token: "function", foreground: "8250df" },
    ],
    colors: {
      "editor.background": "#ffffff",
      "editor.foreground": "#24292f",
      "editor.lineHighlightBackground": "#f6f8fa",
      "editorCursor.foreground": "#24292f",
      "editor.selectionBackground": "#d0d7de",
    },
  });
}

// ── Error Gutter Marker ─────────────────────────────────────────────────────

function setErrorMarkers(
  editor: editor.IStandaloneCodeEditor,
  monacoInstance: typeof monaco,
  errorLines: number[]
) {
  const model = editor.getModel();
  if (!model) return;

  const markers: monaco.editor.IMarkerData[] = errorLines.map((line) => ({
    severity: monacoInstance.MarkerSeverity.Error,
    message: "Error on this line",
    startLineNumber: line,
    startColumn: 1,
    endLineNumber: line,
    endColumn: model.getLineLength(line) + 1 || 1,
  }));

  monacoInstance.editor.setModelMarkers(model, "code-editor", markers);
}

// ── Memoized toolbar button ─────────────────────────────────────────────────

const ToolbarButton = memo(function ToolbarButton({
  onClick,
  title,
  ariaLabel,
  children,
  active,
}: {
  onClick: () => void;
  title: string;
  ariaLabel: string;
  children: React.ReactNode;
  active?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      className={`p-1.5 rounded transition-colors ${
        active
          ? "text-indigo-400 bg-indigo-500/10 hover:bg-indigo-500/20"
          : "text-slate-400 hover:text-white hover:bg-slate-700"
      }`}
      title={title}
      aria-label={ariaLabel}
      style={{ minWidth: "32px", minHeight: "32px" }}
    >
      {children}
    </button>
  );
});

// ── Theme Selector Dropdown ─────────────────────────────────────────────────

const ThemeSelector = memo(function ThemeSelector({
  currentTheme,
  onThemeChange,
}: {
  currentTheme: EditorTheme;
  onThemeChange: (t: EditorTheme) => void;
}) {
  const [open, setOpen] = useState(false);

  return (
    <div className="relative">
      <ToolbarButton
        onClick={() => setOpen(!open)}
        title="Select editor theme"
        ariaLabel="Select editor theme"
      >
        <Palette className="w-4 h-4" />
      </ToolbarButton>
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute left-0 top-full mt-1 w-44 bg-slate-800 border border-slate-600 rounded-lg shadow-xl z-50 py-1">
            {THEME_OPTIONS.map((opt) => (
              <button
                key={opt.id}
                onClick={() => {
                  onThemeChange(opt.id);
                  setOpen(false);
                }}
                className={`w-full text-left px-3 py-1.5 text-xs transition-colors ${
                  currentTheme === opt.id
                    ? "bg-indigo-500/20 text-indigo-300"
                    : "text-slate-300 hover:bg-slate-700"
                }`}
              >
                <span
                  className="inline-block w-3 h-3 rounded-full mr-2 align-middle"
                  style={{
                    backgroundColor:
                      opt.id === "vs-dark" || opt.id === "monokai" || opt.id === "dracula"
                        ? "#1e1e1e"
                        : "#ffffff",
                    border: "1px solid #64748b",
                  }}
                />
                {opt.label}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
});

// ── Editor Toolbar ──────────────────────────────────────────────────────────

interface EditorToolbarProps {
  language: string;
  hasValue: boolean;
  template?: string;
  onSubmit?: () => void;
  onCopy: () => void;
  onReset: () => void;
  currentTheme: EditorTheme;
  onThemeChange: (t: EditorTheme) => void;
  fontSize: number;
  onFontSizeChange: (s: number) => void;
  wordWrap: boolean;
  onWordWrapToggle: () => void;
  minimap: boolean;
  onMinimapToggle: () => void;
  diffMode: boolean;
  onDiffModeToggle: () => void;
}

const EditorToolbar = memo(function EditorToolbar({
  language,
  onSubmit,
  onCopy,
  onReset,
  currentTheme,
  onThemeChange,
  fontSize,
  onFontSizeChange,
  wordWrap,
  onWordWrapToggle,
  minimap,
  onMinimapToggle,
  diffMode,
  onDiffModeToggle,
}: EditorToolbarProps) {
  const [fontMenuOpen, setFontMenuOpen] = useState(false);

  return (
    <div className="flex items-center justify-between px-3 py-1.5 bg-slate-800 border-b border-slate-700">
      {/* Left section */}
      <div className="flex items-center gap-1.5">
        {/* Theme selector */}
        <ThemeSelector currentTheme={currentTheme} onThemeChange={onThemeChange} />

        {/* Font size */}
        <div className="relative">
          <ToolbarButton
            onClick={() => setFontMenuOpen(!fontMenuOpen)}
            title="Font size"
            ariaLabel="Font size"
          >
            <span className="text-[10px] font-mono font-bold">{fontSize}</span>
          </ToolbarButton>
          {fontMenuOpen && (
            <>
              <div className="fixed inset-0 z-40" onClick={() => setFontMenuOpen(false)} />
              <div className="absolute left-0 top-full mt-1 w-20 bg-slate-800 border border-slate-600 rounded-lg shadow-xl z-50 py-1 max-h-48 overflow-y-auto">
                {FONT_SIZES.map((s) => (
                  <button
                    key={s}
                    onClick={() => {
                      onFontSizeChange(s);
                      setFontMenuOpen(false);
                    }}
                    className={`w-full text-center px-2 py-1 text-xs transition-colors ${
                      fontSize === s
                        ? "bg-indigo-500/20 text-indigo-300"
                        : "text-slate-300 hover:bg-slate-700"
                    }`}
                  >
                    {s}
                  </button>
                ))}
              </div>
            </>
          )}
        </div>

        {/* Word wrap */}
        <ToolbarButton
          onClick={onWordWrapToggle}
          title={wordWrap ? "Disable word wrap" : "Enable word wrap"}
          ariaLabel={wordWrap ? "Disable word wrap" : "Enable word wrap"}
          active={wordWrap}
        >
          <AlignLeft className="w-4 h-4" />
        </ToolbarButton>

        {/* Minimap */}
        <ToolbarButton
          onClick={onMinimapToggle}
          title={minimap ? "Hide minimap" : "Show minimap"}
          ariaLabel={minimap ? "Hide minimap" : "Show minimap"}
          active={minimap}
        >
          <Minimize2 className="w-4 h-4" />
        </ToolbarButton>

        {/* Diff mode */}
        <ToolbarButton
          onClick={onDiffModeToggle}
          title="Toggle diff view"
          ariaLabel="Toggle diff view"
          active={diffMode}
        >
          <GitCompare className="w-4 h-4" />
        </ToolbarButton>

        {/* Language badge */}
        <span className="ml-2 text-[10px] font-medium text-slate-500 uppercase tracking-wider hidden sm:inline">
          {language}
        </span>
      </div>

      {/* Right section */}
      <div className="flex items-center gap-1">
        <ToolbarButton onClick={onReset} title="Reset to template" ariaLabel="Reset to template">
          <RotateCcw className="w-4 h-4" />
        </ToolbarButton>
        <ToolbarButton onClick={onCopy} title="Copy code" ariaLabel="Copy code">
          <Copy className="w-4 h-4" />
        </ToolbarButton>
        {onSubmit && (
          <button
            onClick={onSubmit}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-green-600 text-white text-xs font-medium rounded hover:bg-green-700 transition-colors"
            aria-label="Submit solution"
            style={{ minHeight: "32px" }}
          >
            <Play className="w-3.5 h-3.5" />
            <span>Submit</span>
          </button>
        )}
      </div>
    </div>
  );
});

// ── Main CodeEditor Component ───────────────────────────────────────────────

export const CodeEditor = memo(function CodeEditor({
  value,
  onChange,
  language = "go",
  theme: initialTheme = "vs-dark",
  readOnly = false,
  template,
  onSubmit,
  previousSubmission,
  errorLines = [],
  height,
}: CodeEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const monacoRef = useRef<typeof monaco | null>(null);
  const [currentTheme, setCurrentTheme] = useState<EditorTheme>(initialTheme);
  const [fontSize, setFontSize] = useState(14);
  const [wordWrap, setWordWrap] = useState(true);
  const [minimap, setMinimap] = useState(false);
  const [diffMode, setDiffMode] = useState(false);

  // Debounced error lines to avoid excessive marking
  const debouncedErrorLines = useDebounce(errorLines, 300);

  // Handle editor mount
  const handleEditorMount: OnMount = useCallback(
    (editor, monacoInstance) => {
      editorRef.current = editor;
      monacoRef.current = monacoInstance;

      // Define custom themes
      defineCustomThemes(monacoInstance);

      // Register Go completion provider (once)
      const disposable = monacoInstance.languages.registerCompletionItemProvider("go", {
        provideCompletionItems: (model: monaco.editor.ITextModel, position: monaco.Position) => {
          const word = model.getWordUntilPosition(position);
          const range = {
            startLineNumber: position.lineNumber,
            endLineNumber: position.lineNumber,
            startColumn: word.startColumn,
            endColumn: word.endColumn,
          };

          return {
            suggestions: GO_SNIPPETS.map((s) => ({ ...s, range })),
          };
        },
        triggerCharacters: [".", " "],
      });

      // Register Go formatting on save
      monacoInstance.languages.registerDocumentFormattingEditProvider("go", {
        provideDocumentFormattingEdits: (model: monaco.editor.ITextModel) => {
          // Simple gofmt-like formatting: replace tabs, ensure consistent indentation
          const fullText = model.getValue();
          const lines = fullText.split("\n");
          const formatted: string = lines
            .map((line: string) => {
              // Trim trailing whitespace
              const trimmed = line.replace(/\s+$/, "");
              // Ensure tabs for indentation (4-space -> tab)
              return trimmed.replace(/^ {4}/gm, "\t");
            })
            .join("\n");

          if (formatted === fullText) return [];

          return [
            {
              range: model.getFullModelRange(),
              text: formatted,
            },
          ];
        },
      });

      // Register Go import auto-completion
      monacoInstance.languages.registerCompletionItemProvider("go", {
        provideCompletionItems: (model: monaco.editor.ITextModel, position: monaco.Position) => {
          const textUntilPosition = model.getValueInRange({
            startLineNumber: position.lineNumber,
            startColumn: 1,
            endLineNumber: position.lineNumber,
            endColumn: position.column,
          });

          // Only trigger inside import block
          const importMatch = textUntilPosition.match(/^\s*(?:")([^"]*)$/);
          if (!importMatch) return { suggestions: [] };

          const commonImports = [
            "fmt",
            "strings",
            "strconv",
            "math",
            "sort",
            "os",
            "io",
            "bufio",
            "errors",
            "sync",
            "time",
            "encoding/json",
            "encoding/base64",
            "net/http",
            "regexp",
            "container/list",
            "container/heap",
            "sort",
            "unicode",
            "bytes",
            "reflect",
            "path/filepath",
            "crypto/sha256",
            "crypto/md5",
            "flag",
            "log",
          ];

          const word = model.getWordUntilPosition(position);
          const range = {
            startLineNumber: position.lineNumber,
            endLineNumber: position.lineNumber,
            startColumn: word.startColumn,
            endColumn: word.endColumn,
          };

          return {
            suggestions: commonImports.map((pkg) => ({
              label: pkg,
              kind: monacoInstance.languages.CompletionItemKind.Module,
              insertText: pkg,
              detail: "Go standard library",
              range,
            })),
          };
        },
      });

      // Keyboard shortcut: Ctrl+Enter to submit
      editor.addCommand(monacoInstance.KeyMod.CtrlCmd | monacoInstance.KeyCode.Enter, () => {
        onSubmit?.();
      });

      return () => {
        disposable.dispose();
      };
    },
    [onSubmit]
  );

  // Set error markers when they change
  useEffect(() => {
    if (editorRef.current && monacoRef.current && debouncedErrorLines.length > 0) {
      setErrorMarkers(editorRef.current, monacoRef.current, debouncedErrorLines);
    }
    if (editorRef.current && monacoRef.current && debouncedErrorLines.length === 0) {
      monacoRef.current.editor.setModelMarkers(
        editorRef.current.getModel()!,
        "code-editor",
        []
      );
    }
  }, [debouncedErrorLines]);

  // Handlers
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(value).catch(() => {});
  }, [value]);

  const handleReset = useCallback(() => {
    if (template) onChange(template);
  }, [template, onChange]);

  const handleThemeChange = useCallback((t: EditorTheme) => {
    setCurrentTheme(t);
  }, []);

  const handleFontSizeChange = useCallback((s: number) => {
    setFontSize(s);
  }, []);

  const handleWordWrapToggle = useCallback(() => {
    setWordWrap((p) => !p);
  }, []);

  const handleMinimapToggle = useCallback(() => {
    setMinimap((p) => !p);
  }, []);

  const handleDiffModeToggle = useCallback(() => {
    setDiffMode((p) => !p);
  }, []);

  // Editor options
  const editorOptions = useMemo(
    () => ({
      readOnly,
      minimap: { enabled: minimap },
      fontSize,
      lineNumbers: "on" as const,
      roundedSelection: true,
      scrollBeyondLastLine: false,
      automaticLayout: true,
      tabSize: 4,
      wordWrap: (wordWrap ? "on" : "off") as "on" | "off",
      padding: { top: 12, bottom: 12 },
      smoothScrolling: true,
      cursorBlinking: "smooth" as const,
      cursorSmoothCaretAnimation: "on" as const,
      renderLineHighlight: "all" as const,
      suggestOnTriggerCharacters: true,
      acceptSuggestionOnEnter: "on" as const,
      bracketPairColorization: { enabled: true },
      guides: { indentation: true, bracketPairs: true },
      folding: true,
      foldingHighlight: true,
      links: true,
      formatOnPaste: true,
      formatOnType: true,
      suggest: {
        snippetsPreventQuickSuggestions: false,
        showKeywords: true,
        showSnippets: true,
      },
    }),
    [readOnly, minimap, fontSize, wordWrap]
  );

  // Insert template on mount
  useEffect(() => {
    if (template && !value && editorRef.current) {
      editorRef.current.setValue(template);
    }
  }, [template, value]);

  const handleChange = useCallback(
    (val: string | undefined) => {
      onChange(val ?? "");
    },
    [onChange]
  );

  // Resolve monaco theme from our EditorTheme
  const resolvedMonacoTheme = currentTheme;

  return (
    <div className="flex flex-col h-full rounded-lg overflow-hidden border border-slate-700">
      <EditorToolbar
        language={language}
        hasValue={!!value}
        template={template}
        onSubmit={onSubmit}
        onCopy={handleCopy}
        onReset={handleReset}
        currentTheme={currentTheme}
        onThemeChange={handleThemeChange}
        fontSize={fontSize}
        onFontSizeChange={handleFontSizeChange}
        wordWrap={wordWrap}
        onWordWrapToggle={handleWordWrapToggle}
        minimap={minimap}
        onMinimapToggle={handleMinimapToggle}
        diffMode={diffMode}
        onDiffModeToggle={handleDiffModeToggle}
      />
      <div className="flex-1 min-h-[300px]" style={{ height: height || "100%" }}>
        {diffMode && previousSubmission ? (
          <DiffEditor
            height="100%"
            language={language}
            original={previousSubmission}
            modified={value}
            theme={resolvedMonacoTheme}
            onMount={(_editor: any, monacoInstance: any) => { defineCustomThemes(monacoInstance); }}
            options={{ ...editorOptions, readOnly: true, renderSideBySide: true, originalEditable: false }}
            loading={<div className="flex items-center justify-center h-full bg-slate-900 text-slate-400">Loading diff...</div>}
          />
        ) : (
          <Editor
            height="100%"
            language={language}
            value={value}
            theme={resolvedMonacoTheme}
            onChange={handleChange}
            onMount={handleEditorMount}
            options={editorOptions}
            loading={
              <div className="flex items-center justify-center h-full bg-slate-900 text-slate-400">
                Loading editor...
              </div>
            }
          />
        )}
      </div>
    </div>
  );
});

export default CodeEditor;
