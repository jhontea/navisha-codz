import React, { useState, useCallback, memo } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
  ArrowLeft,
  Save,
  Plus,
  Trash2,
  Eye,
  EyeOff,
  ToggleLeft,
  ToggleRight,
  GripVertical,
  ChevronDown,
  ChevronUp,
  Lightbulb,
  TestTube,
  Code2,
  FileText,
  AlertCircle,
  Check,
} from "lucide-react";
import Editor, { OnMount } from "@monaco-editor/react";
import type { editor } from "monaco-editor";
import * as monaco from "monaco-editor";
import type { Difficulty, Category, TestCase, Hint } from "../../types";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { CATEGORY_LIST } from "../../components/problem/CategoryIcon";

// ── Types ───────────────────────────────────────────────────────────────────

interface ProblemFormData {
  title: string;
  slug: string;
  difficulty: Difficulty;
  category: Category;
  description: string;
  points: number;
  time_limit_ms: number;
  memory_limit_mb: number;
  function_template: string;
  solution_code: string;
  is_published: boolean;
  tags: string[];
  constraints: string[];
  examples: Array<{ input: string; output: string; explanation: string }>;
  test_cases: TestCase[];
  hints: HintInput[];
}

interface HintInput {
  id: string;
  level: number;
  content: string;
  penalty_points: number;
}

const EMPTY_FORM: ProblemFormData = {
  title: "",
  slug: "",
  difficulty: "easy",
  category: "arrays" as Category,
  description: "",
  points: 100,
  time_limit_ms: 1000,
  memory_limit_mb: 256,
  function_template: "",
  solution_code: "",
  is_published: false,
  tags: [],
  constraints: [],
  examples: [{ input: "", output: "", explanation: "" }],
  test_cases: [],
  hints: [],
};

const DIFFICULTIES: Difficulty[] = ["easy", "medium", "hard"];

// ── Test Case Row ───────────────────────────────────────────────────────────

const TestCaseRow = memo(function TestCaseRow({
  index,
  testCase,
  onChange,
  onRemove,
}: {
  index: number;
  testCase: Partial<TestCase>;
  onChange: (idx: number, field: string, value: string | boolean) => void;
  onRemove: (idx: number) => void;
}) {
  return (
    <div className="group relative border border-slate-200 dark:border-slate-700 rounded-lg p-4 space-y-3 hover:border-slate-300 dark:hover:border-slate-600 transition-colors">
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
          Test Case #{index + 1}
        </span>
        <button
          onClick={() => onRemove(index)}
          className="p-1 rounded text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/30 transition-colors opacity-0 group-hover:opacity-100"
          aria-label={`Remove test case ${index + 1}`}
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div>
          <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Input</label>
          <textarea
            value={testCase.input || ""}
            onChange={(e) => onChange(index, "input", e.target.value)}
            className="w-full px-3 py-2 text-xs font-mono border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            rows={2}
            placeholder="Test input..."
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Expected Output</label>
          <textarea
            value={testCase.expected_output || ""}
            onChange={(e) => onChange(index, "expected_output", e.target.value)}
            className="w-full px-3 py-2 text-xs font-mono border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            rows={2}
            placeholder="Expected output..."
          />
        </div>
      </div>
      <div>
        <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Description (optional)</label>
        <input
          value={testCase.description || ""}
          onChange={(e) => onChange(index, "description", e.target.value)}
          className="w-full px-3 py-1.5 text-xs border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          placeholder="e.g. Basic test case"
        />
      </div>
    </div>
  );
});

// ── Hint Row ────────────────────────────────────────────────────────────────

const HintRow = memo(function HintRow({
  index,
  hint,
  onChange,
  onRemove,
}: {
  index: number;
  hint: HintInput;
  onChange: (idx: number, field: string, value: string | number) => void;
  onRemove: (idx: number) => void;
}) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between px-4 py-3 bg-slate-50 dark:bg-slate-800/50 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
      >
        <div className="flex items-center gap-2">
          <Lightbulb className="w-4 h-4 text-amber-500" />
          <span className="text-sm font-medium text-slate-700 dark:text-slate-300">
            Hint Level {hint.level}
          </span>
          {hint.content && (
            <span className="text-xs text-slate-400 dark:text-slate-500 ml-2">
              ({hint.content.length} chars)
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-slate-400">-{hint.penalty_points} pts</span>
          {expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
        </div>
      </button>
      {expanded && (
        <div className="p-4 space-y-3 border-t border-slate-200 dark:border-slate-700">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Level</label>
              <select
                value={hint.level}
                onChange={(e) => onChange(index, "level", parseInt(e.target.value))}
                className="w-full px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200"
              >
                {[1, 2, 3].map((l) => (
                  <option key={l} value={l}>
                    Level {l}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Penalty Points</label>
              <input
                type="number"
                min={0}
                value={hint.penalty_points}
                onChange={(e) => onChange(index, "penalty_points", parseInt(e.target.value) || 0)}
                className="w-full px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200"
              />
            </div>
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Hint Content</label>
            <textarea
              value={hint.content}
              onChange={(e) => onChange(index, "content", e.target.value)}
              className="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 focus:ring-2 focus:ring-indigo-500"
              rows={3}
              placeholder="Enter hint content..."
            />
          </div>
          <button
            onClick={() => onRemove(index)}
            className="flex items-center gap-1 text-xs text-red-500 hover:text-red-600 transition-colors"
          >
            <Trash2 className="w-3 h-3" />
            Remove hint
          </button>
        </div>
      )}
    </div>
  );
});

// ── Main Component ──────────────────────────────────────────────────────────

export function AdminProblemForm() {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const isEditing = Boolean(slug);

  const [form, setForm] = useState<ProblemFormData>(EMPTY_FORM);
  const [showPreview, setShowPreview] = useState(false);
  const [showSolution, setShowSolution] = useState(false);
  const [saving, setSaving] = useState(false);
  const [newTag, setNewTag] = useState("");
  const [newConstraint, setNewConstraint] = useState("");

  // ── Field updater ────────────────────────────────────────────────────────

  const updateField = useCallback(
    <K extends keyof ProblemFormData>(key: K, value: ProblemFormData[K]) => {
      setForm((prev) => ({ ...prev, [key]: value }));
    },
    []
  );

  // ── Test cases ───────────────────────────────────────────────────────────

  const addTestCase = useCallback(() => {
    setForm((prev) => ({
      ...prev,
      test_cases: [
        ...prev.test_cases,
        {
          id: `tc-${Date.now()}`,
          input: "",
          expected_output: "",
          is_sample: false,
        },
      ],
    }));
  }, []);

  const updateTestCase = useCallback(
    (idx: number, field: string, value: string | boolean) => {
      setForm((prev) => {
        const updated = [...prev.test_cases];
        updated[idx] = { ...updated[idx], [field]: value };
        return { ...prev, test_cases: updated };
      });
    },
    []
  );

  const removeTestCase = useCallback((idx: number) => {
    setForm((prev) => ({
      ...prev,
      test_cases: prev.test_cases.filter((_, i) => i !== idx),
    }));
  }, []);

  // ── Hints ────────────────────────────────────────────────────────────────

  const addHint = useCallback(() => {
    setForm((prev) => ({
      ...prev,
      hints: [
        ...prev.hints,
        {
          id: `hint-${Date.now()}`,
          level: prev.hints.length + 1,
          content: "",
          penalty_points: 10 * (prev.hints.length + 1),
        },
      ],
    }));
  }, []);

  const updateHint = useCallback(
    (idx: number, field: string, value: string | number) => {
      setForm((prev) => {
        const updated = [...prev.hints];
        updated[idx] = { ...updated[idx], [field]: value };
        return { ...prev, hints: updated };
      });
    },
    []
  );

  const removeHint = useCallback((idx: number) => {
    setForm((prev) => ({
      ...prev,
      hints: prev.hints.filter((_, i) => i !== idx),
    }));
  }, []);

  // ── Tags & Constraints ───────────────────────────────────────────────────

  const addTag = useCallback(() => {
    const tag = newTag.trim();
    if (tag && !form.tags.includes(tag)) {
      setForm((prev) => ({ ...prev, tags: [...prev.tags, tag] }));
    }
    setNewTag("");
  }, [newTag, form.tags]);

  const removeTag = useCallback((idx: number) => {
    setForm((prev) => ({
      ...prev,
      tags: prev.tags.filter((_, i) => i !== idx),
    }));
  }, []);

  const addConstraint = useCallback(() => {
    const c = newConstraint.trim();
    if (c) {
      setForm((prev) => ({ ...prev, constraints: [...prev.constraints, c] }));
    }
    setNewConstraint("");
  }, [newConstraint, form.constraints]);

  const removeConstraint = useCallback((idx: number) => {
    setForm((prev) => ({
      ...prev,
      constraints: prev.constraints.filter((_, i) => i !== idx),
    }));
  }, []);

  // ── Examples ─────────────────────────────────────────────────────────────

  const addExample = useCallback(() => {
    setForm((prev) => ({
      ...prev,
      examples: [...prev.examples, { input: "", output: "", explanation: "" }],
    }));
  }, []);

  const updateExample = useCallback(
    (idx: number, field: string, value: string) => {
      setForm((prev) => {
        const updated = [...prev.examples];
        updated[idx] = { ...updated[idx], [field]: value };
        return { ...prev, examples: updated };
      });
    },
    []
  );

  const removeExample = useCallback((idx: number) => {
    setForm((prev) => ({
      ...prev,
      examples: prev.examples.filter((_, i) => i !== idx),
    }));
  }, []);

  // ── Submit ───────────────────────────────────────────────────────────────

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setSaving(true);
      try {
        // Simulate save — in production, call API
        await new Promise((resolve) => setTimeout(resolve, 1000));
        navigate("/admin");
      } catch (err) {
        console.error("Failed to save problem:", err);
      } finally {
        setSaving(false);
      }
    },
    [form, navigate]
  );

  // ── Slug auto-generation ─────────────────────────────────────────────────

  const generateSlug = useCallback(() => {
    const slug = form.title
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "");
    updateField("slug", slug);
  }, [form.title, updateField]);

  // ── Render markdown preview ──────────────────────────────────────────────

  const renderMarkdownPreview = (text: string) => {
    if (!text) return <span className="text-slate-400 italic">No content yet...</span>;
    return (
      <div className="prose prose-sm dark:prose-invert max-w-none whitespace-pre-wrap">
        {text.split("\n").map((line, i) => {
          if (line.startsWith("## ")) return <h2 key={i} className="text-lg font-bold mt-4 mb-2">{line.slice(3)}</h2>;
          if (line.startsWith("### ")) return <h3 key={i} className="text-base font-semibold mt-3 mb-1">{line.slice(4)}</h3>;
          if (line.startsWith("**") && line.endsWith("**")) return <p key={i} className="font-semibold">{line.slice(2, -2)}</p>;
          if (line.startsWith("- ")) return <li key={i} className="ml-4 list-disc text-sm">{line.slice(2)}</li>;
          if (line.trim() === "") return <br key={i} />;
          return <p key={i} className="text-sm mb-1">{line}</p>;
        })}
      </div>
    );
  };

  // ── Render ───────────────────────────────────────────────────────────────

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <Link
            to="/admin"
            className="inline-flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 mb-2"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to Dashboard
          </Link>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">
            {isEditing ? "Edit Problem" : "Create New Problem"}
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
            {isEditing ? `Editing: ${slug}` : "Define a new coding challenge"}
          </p>
        </div>
        <div className="flex items-center gap-3">
          {/* Publish/Draft Toggle */}
          <button
            onClick={() => updateField("is_published", !form.is_published)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              form.is_published
                ? "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                : "bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400"
            }`}
            aria-label={form.is_published ? "Published" : "Draft"}
          >
            {form.is_published ? (
              <ToggleRight className="w-5 h-5" />
            ) : (
              <ToggleLeft className="w-5 h-5" />
            )}
            {form.is_published ? "Published" : "Draft"}
          </button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            loading={saving}
            icon={<Save className="w-4 h-4" />}
          >
            {isEditing ? "Update Problem" : "Create Problem"}
          </Button>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-8">
        {/* ── Basic Info ─────────────────────────────────────────────────────── */}
        <section className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
          <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
            <FileText className="w-5 h-5 text-indigo-500" />
            Basic Information
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Title *</label>
              <input
                type="text"
                value={form.title}
                onChange={(e) => {
                  updateField("title", e.target.value);
                  if (!isEditing && !form.slug) {
                    const autoSlug = e.target.value
                      .toLowerCase()
                      .replace(/[^a-z0-9]+/g, "-")
                      .replace(/^-+|-+$/g, "");
                    updateField("slug", autoSlug);
                  }
                }}
                className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-indigo-500"
                placeholder="e.g. Two Sum"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Slug *</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={form.slug}
                  onChange={(e) => updateField("slug", e.target.value)}
                  className="flex-1 px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-indigo-500"
                  placeholder="two-sum"
                  required
                />
                <button
                  type="button"
                  onClick={generateSlug}
                  className="px-3 py-2 text-xs font-medium text-indigo-600 bg-indigo-50 dark:bg-indigo-950/30 rounded-lg hover:bg-indigo-100 dark:hover:bg-indigo-950/50 transition-colors"
                >
                  Auto
                </button>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Difficulty</label>
              <select
                value={form.difficulty}
                onChange={(e) => updateField("difficulty", e.target.value as Difficulty)}
                className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
              >
                {DIFFICULTIES.map((d) => (
                  <option key={d} value={d}>
                    {d.charAt(0).toUpperCase() + d.slice(1)}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Category</label>
              <select
                value={form.category}
                onChange={(e) => updateField("category", e.target.value as Category)}
                className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
              >
                {CATEGORY_LIST.map((cat) => (
                  <option key={cat.value} value={cat.value}>
                    {cat.label}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Points</label>
              <input
                type="number"
                min={0}
                value={form.points}
                onChange={(e) => updateField("points", parseInt(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Time Limit (ms)</label>
                <input
                  type="number"
                  min={100}
                  step={100}
                  value={form.time_limit_ms}
                  onChange={(e) => updateField("time_limit_ms", parseInt(e.target.value) || 1000)}
                  className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Memory Limit (MB)</label>
                <input
                  type="number"
                  min={16}
                  step={16}
                  value={form.memory_limit_mb}
                  onChange={(e) => updateField("memory_limit_mb", parseInt(e.target.value) || 256)}
                  className="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                />
              </div>
            </div>
          </div>
        </section>

        {/* ── Description + Markdown Preview ─────────────────────────────────── */}
        <section className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <FileText className="w-5 h-5 text-indigo-500" />
              Description
            </h2>
            <button
              type="button"
              onClick={() => setShowPreview(!showPreview)}
              className={`flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors ${
                showPreview
                  ? "bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400"
                  : "text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
              }`}
            >
              {showPreview ? <Code2 className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
              {showPreview ? "Edit" : "Preview"}
            </button>
          </div>
          {showPreview ? (
            <div className="p-4 border border-slate-200 dark:border-slate-700 rounded-lg bg-slate-50 dark:bg-slate-800/50 min-h-[200px]">
              {renderMarkdownPreview(form.description)}
            </div>
          ) : (
            <textarea
              value={form.description}
              onChange={(e) => updateField("description", e.target.value)}
              className="w-full px-4 py-3 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-indigo-500 font-mono text-sm min-h-[200px]"
              placeholder="Describe the problem... Supports ## headings, **bold**, - lists"
            />
          )}
        </section>

        {/* ── Examples ────────────────────────────────────────────────────────── */}
        <section className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <TestTube className="w-5 h-5 text-emerald-500" />
              Examples
            </h2>
            <button
              type="button"
              onClick={addExample}
              className="flex items-center gap-1 px-3 py-1.5 text-xs font-medium text-indigo-600 bg-indigo-50 dark:bg-indigo-950/30 rounded-lg hover:bg-indigo-100 dark:hover:bg-indigo-950/50 transition-colors"
            >
              <Plus className="w-3.5 h-3.5" />
              Add Example
            </button>
          </div>
          <div className="space-y-3">
            {form.examples.map((ex, idx) => (
              <div key={idx} className="border border-slate-200 dark:border-slate-700 rounded-lg p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold text-slate-500">Example #{idx + 1}</span>
                  {form.examples.length > 1 && (
                    <button
                      type="button"
                      onClick={() => removeExample(idx)}
                      className="p-1 rounded text-slate-400 hover:text-red-500 transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  )}
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-slate-500 mb-1">Input</label>
                    <textarea
                      value={ex.input}
                      onChange={(e) => updateExample(idx, "input", e.target.value)}
                      className="w-full px-3 py-2 text-sm font-mono border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                      rows={2}
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-slate-500 mb-1">Output</label>
                    <textarea
                      value={ex.output}
                      onChange={(e) => updateExample(idx, "output", e.target.value)}
                      className="w-full px-3 py-2 text-sm font-mono border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                      rows={2}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-500 mb-1">Explanation (optional)</label>
                  <textarea
                    value={ex.explanation}
                    onChange={(e) => updateExample(idx, "explanation", e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                    rows={2}
                  />
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* ── Constraints + Tags ──────────────────────────────────────────────── */}
        <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Constraints */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <AlertCircle className="w-5 h-5 text-amber-500" />
              Constraints
            </h2>
            <div className="flex gap-2">
              <input
                type="text"
                value={newConstraint}
                onChange={(e) => setNewConstraint(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), addConstraint())}
                className="flex-1 px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                placeholder="e.g. 1 <= n <= 10^5"
              />
              <button
                type="button"
                onClick={addConstraint}
                className="px-3 py-1.5 text-sm font-medium text-indigo-600 bg-indigo-50 dark:bg-indigo-950/30 rounded-lg hover:bg-indigo-100 transition-colors"
              >
                Add
              </button>
            </div>
            <div className="space-y-1">
              {form.constraints.map((c, idx) => (
                <div key={idx} className="flex items-center justify-between px-3 py-1.5 bg-slate-50 dark:bg-slate-800/50 rounded-lg text-sm text-slate-700 dark:text-slate-300">
                  <span className="font-mono text-xs">{c}</span>
                  <button type="button" onClick={() => removeConstraint(idx)} className="text-slate-400 hover:text-red-500">
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>
              ))}
              {form.constraints.length === 0 && (
                <p className="text-xs text-slate-400 italic">No constraints added yet</p>
              )}
            </div>
          </div>

          {/* Tags */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <Code2 className="w-5 h-5 text-cyan-500" />
              Tags
            </h2>
            <div className="flex gap-2">
              <input
                type="text"
                value={newTag}
                onChange={(e) => setNewTag(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), addTag())}
                className="flex-1 px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
                placeholder="e.g. hash-table"
              />
              <button
                type="button"
                onClick={addTag}
                className="px-3 py-1.5 text-sm font-medium text-indigo-600 bg-indigo-50 dark:bg-indigo-950/30 rounded-lg hover:bg-indigo-100 transition-colors"
              >
                Add
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {form.tags.map((tag, idx) => (
                <span
                  key={idx}
                  className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-medium rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300"
                >
                  {tag}
                  <button type="button" onClick={() => removeTag(idx)} className="hover:text-red-500">&times;</button>
                </span>
              ))}
              {form.tags.length === 0 && (
                <p className="text-xs text-slate-400 italic">No tags added yet</p>
              )}
            </div>
          </div>
        </section>

        {/* ── Test Cases ───────────────────────────────────────────────────────── */}
        <section className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <TestTube className="w-5 h-5 text-emerald-500" />
              Test Cases
            </h2>
            <button
              type="button"
              onClick={addTestCase}
              className="flex items-center gap-1 px-3 py-1.5 text-xs font-medium text-indigo-600 bg-indigo-50 dark:bg-indigo-950/30 rounded-lg hover:bg-indigo-100 transition-colors"
            >
              <Plus className="w-3.5 h-3.5" />
              Add Test Case
            </button>
          </div>
          <div className="space-y-3">
            {form.test_cases.length === 0 ? (
              <p className="text-sm text-slate-400 italic text-center py-6">
                No test cases added yet. Click "Add Test Case" to create one.
              </p>
            ) : (
              form.test_cases.map((tc, idx) => (
                <TestCaseRow
                  key={tc.id}
                  index={idx}
                  testCase={tc}
                  onChange={updateTestCase}
                  onRemove={removeTestCase}
                />
              ))
            )}
          </div>
        </section>

        {/* ── Hints ────────────────────────────────────────────────────────────── */}
        <section className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <Lightbulb className="w-5 h-5 text-amber-500" />
              Hints (3 Levels)
            </h2>
            <button
              type="button"
              onClick={addHint}
              className="flex items-center gap-1 px-3 py-1.5 text-xs font-medium text-indigo-600 bg-indigo-50 dark:bg-indigo-950/30 rounded-lg hover:bg-indigo-100 transition-colors"
            >
              <Plus className="w-3.5 h-3.5" />
              Add Hint
            </button>
          </div>
          <div className="space-y-2">
            {form.hints.length === 0 ? (
              <p className="text-sm text-slate-400 italic text-center py-6">
                No hints yet. Add up to 3 hints with increasing levels.
              </p>
            ) : (
              form.hints.map((hint, idx) => (
                <HintRow
                  key={hint.id}
                  index={idx}
                  hint={hint}
                  onChange={updateHint}
                  onRemove={removeHint}
                />
              ))
            )}
          </div>
        </section>

        {/* ── Code Editors ─────────────────────────────────────────────────────── */}
        <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Template Code */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-3">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <Code2 className="w-5 h-5 text-indigo-500" />
              Function Template
            </h2>
            <p className="text-xs text-slate-500">Code template shown to the user when starting the problem.</p>
            <div className="h-[300px] border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden">
              <Editor
                height="100%"
                defaultLanguage="go"
                value={form.function_template}
                onChange={(val) => updateField("function_template", val ?? "")}
                theme="vs-dark"
                options={{
                  minimap: { enabled: false },
                  fontSize: 13,
                  lineNumbers: "on",
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  tabSize: 4,
                  wordWrap: "on",
                  padding: { top: 8, bottom: 8 },
                }}
              />
            </div>
          </div>

          {/* Solution Code (Hidden) */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-3">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
                <Code2 className="w-5 h-5 text-rose-500" />
                Solution
              </h2>
              <button
                type="button"
                onClick={() => setShowSolution(!showSolution)}
                className={`flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors ${
                  showSolution
                    ? "bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400"
                    : "text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
                }`}
              >
                {showSolution ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                {showSolution ? "Hide" : "Show"}
              </button>
            </div>
            <p className="text-xs text-slate-500">
              {showSolution ? "Solution is visible. Be careful!" : "Hidden solution — click Show to reveal."}
            </p>
            <div className="h-[300px] border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden relative">
              {!showSolution && (
                <div className="absolute inset-0 bg-slate-900/80 backdrop-blur-sm flex items-center justify-center z-10 rounded-lg">
                  <p className="text-slate-400 text-sm flex items-center gap-2">
                    <EyeOff className="w-4 h-4" />
                    Solution hidden
                  </p>
                </div>
              )}
              <Editor
                height="100%"
                defaultLanguage="go"
                value={form.solution_code}
                onChange={(val) => updateField("solution_code", val ?? "")}
                theme={showSolution ? "vs-dark" : "vs-dark"}
                options={{
                  minimap: { enabled: false },
                  fontSize: 13,
                  lineNumbers: "on",
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  tabSize: 4,
                  wordWrap: "on",
                  padding: { top: 8, bottom: 8 },
                  readOnly: !showSolution,
                }}
              />
            </div>
          </div>
        </section>

        {/* ── Bottom Submit ────────────────────────────────────────────────────── */}
        <div className="flex items-center justify-between bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
          <div className="flex items-center gap-3 text-sm text-slate-500">
            <span>{form.test_cases.length} test case(s)</span>
            <span>{form.hints.length} hint(s)</span>
            <span>{form.tags.length} tag(s)</span>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/admin"
              className="px-4 py-2 text-sm font-medium text-slate-600 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200 transition-colors"
            >
              Cancel
            </Link>
            <Button
              variant="primary"
              onClick={handleSubmit}
              loading={saving}
              icon={<Save className="w-4 h-4" />}
            >
              {isEditing ? "Update Problem" : "Create Problem"}
            </Button>
          </div>
        </div>
      </form>
    </div>
  );
}

export default AdminProblemForm;
