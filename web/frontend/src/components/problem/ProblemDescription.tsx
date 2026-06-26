import React from 'react';
import { Badge } from '../ui/Badge';

export interface ExampleBlock {
  input: string;
  output: string;
  explanation?: string;
}

export interface Constraint {
  label: string;
  value: string;
}

export interface ProblemDescriptionProps {
  title: string;
  difficulty: 'easy' | 'medium' | 'hard' | 'expert';
  description: string;
  examples?: ExampleBlock[];
  constraints?: Constraint[];
  hints?: string[];
  followUp?: string;
  className?: string;
}

export const ProblemDescription: React.FC<ProblemDescriptionProps> = ({
  title,
  difficulty,
  description,
  examples = [],
  constraints = [],
  hints = [],
  followUp,
  className = '',
}) => {
  // Simple markdown-like rendering (bold, code, lists)
  const renderMarkdown = (text: string) => {
    const lines = text.split('\n');
    return lines.map((line, i) => {
      // Bold text
      if (line.startsWith('**') && line.endsWith('**')) {
        return (
          <p key={i} className="font-semibold text-neutral-900 dark:text-neutral-100 mb-2">
            {line.replace(/\*\*/g, '')}
          </p>
        );
      }
      // List items
      if (line.startsWith('- ') || line.startsWith('* ')) {
        return (
          <li key={i} className="ml-4 text-neutral-700 dark:text-neutral-300 list-disc">
            {renderInlineCode(line.slice(2))}
          </li>
        );
      }
      // Numbered list
      if (/^\d+\.\s/.test(line)) {
        return (
          <li key={i} className="ml-4 text-neutral-700 dark:text-neutral-300 list-decimal">
            {renderInlineCode(line.replace(/^\d+\.\s/, ''))}
          </li>
        );
      }
      // Empty line
      if (line.trim() === '') {
        return <br key={i} />;
      }
      // Regular paragraph
      return (
        <p key={i} className="text-neutral-700 dark:text-neutral-300 mb-2">
          {renderInlineCode(line)}
        </p>
      );
    });
  };

  // Render inline code and bold text
  const renderInlineCode = (text: string): React.ReactNode => {
    const parts = text.split(/(`[^`]+`|\*\*[^*]+\*\*)/g);
    return parts.map((part, i) => {
      if (part.startsWith('`') && part.endsWith('`')) {
        return (
          <code key={i} className="px-1.5 py-0.5 text-sm font-mono rounded bg-neutral-100 dark:bg-neutral-800 text-primary-600 dark:text-primary-400">
            {part.slice(1, -1)}
          </code>
        );
      }
      if (part.startsWith('**') && part.endsWith('**')) {
        return <strong key={i} className="font-semibold">{part.slice(2, -2)}</strong>;
      }
      return part;
    });
  };

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header */}
      <div>
        <div className="flex items-center gap-3 mb-2">
          <h1 className="text-2xl font-bold text-neutral-900 dark:text-neutral-100">
            {title}
          </h1>
          <Badge variant="difficulty" difficulty={difficulty} size="md" />
        </div>
      </div>

      {/* Description */}
      <div className="prose-sm">
        {renderMarkdown(description)}
      </div>

      {/* Examples */}
      {examples.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold text-neutral-900 dark:text-neutral-100">
            Examples
          </h2>
          {examples.map((example, index) => (
            <div
              key={index}
              className="p-4 rounded-lg border border-neutral-200 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-800/50"
            >
              <h3 className="text-sm font-semibold text-neutral-900 dark:text-neutral-100 mb-2">
                Example {index + 1}
              </h3>
              <div className="space-y-2">
                <div>
                  <span className="text-xs font-medium text-neutral-500 dark:text-neutral-400">Input: </span>
                  <code className="text-sm font-mono text-neutral-900 dark:text-neutral-100 bg-neutral-100 dark:bg-neutral-800 px-2 py-0.5 rounded">
                    {example.input}
                  </code>
                </div>
                <div>
                  <span className="text-xs font-medium text-neutral-500 dark:text-neutral-400">Output: </span>
                  <code className="text-sm font-mono text-neutral-900 dark:text-neutral-100 bg-neutral-100 dark:bg-neutral-800 px-2 py-0.5 rounded">
                    {example.output}
                  </code>
                </div>
                {example.explanation && (
                  <div className="mt-2 pt-2 border-t border-neutral-200 dark:border-neutral-700">
                    <span className="text-xs font-medium text-neutral-500 dark:text-neutral-400">Explanation: </span>
                    <span className="text-sm text-neutral-700 dark:text-neutral-300">
                      {example.explanation}
                    </span>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Constraints */}
      {constraints.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-neutral-900 dark:text-neutral-100 mb-3">
            Constraints
          </h2>
          <ul className="space-y-1.5">
            {constraints.map((constraint, index) => (
              <li
                key={index}
                className="flex items-start gap-2 text-sm text-neutral-700 dark:text-neutral-300"
              >
                <svg className="w-4 h-4 mt-0.5 text-success-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                <span>
                  <code className="font-mono text-xs bg-neutral-100 dark:bg-neutral-800 px-1.5 py-0.5 rounded text-primary-600 dark:text-primary-400">
                    {constraint.label}
                  </code>
                  <span className="ml-2">{constraint.value}</span>
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Follow Up */}
      {followUp && (
        <div className="p-4 rounded-lg bg-secondary-50 dark:bg-secondary-950/30 border border-secondary-200 dark:border-secondary-800">
          <h3 className="text-sm font-semibold text-secondary-700 dark:text-secondary-300 mb-1">
            Follow-up
          </h3>
          <p className="text-sm text-secondary-600 dark:text-secondary-400">
            {followUp}
          </p>
        </div>
      )}

      {/* Hints */}
      {hints.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-neutral-900 dark:text-neutral-100 mb-3">
            Hints
          </h2>
          <div className="space-y-2">
            {hints.map((hint, index) => (
              <details
                key={index}
                className="group p-3 rounded-lg border border-neutral-200 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-800/50"
              >
                <summary className="cursor-pointer text-sm font-medium text-neutral-700 dark:text-neutral-300 hover:text-primary-600 dark:hover:text-primary-400">
                  Hint {index + 1}
                </summary>
                <p className="mt-2 text-sm text-neutral-600 dark:text-neutral-400">
                  {hint}
                </p>
              </details>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

ProblemDescription.displayName = 'ProblemDescription';

export default ProblemDescription;
