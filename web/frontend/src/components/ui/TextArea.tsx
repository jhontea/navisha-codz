import React, { forwardRef, useRef, useEffect, useCallback } from 'react';

export interface TextAreaProps {
  label?: string;
  error?: string;
  hint?: string;
  size?: 'sm' | 'md' | 'lg';
  autoResize?: boolean;
  showCharCount?: boolean;
  fullWidth?: boolean;
  className?: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  placeholder?: string;
  maxLength?: number;
  id?: string;
  disabled?: boolean;
  rows?: number;
  name?: string;
}

const sizeClasses: Record<string, string> = {
  sm: 'px-3 py-1.5 text-sm',
  md: 'px-3.5 py-2 text-base',
  lg: 'px-4 py-3 text-lg',
};

export const TextArea = forwardRef<HTMLTextAreaElement, TextAreaProps>(
  (
    {
      label,
      error,
      hint,
      size = 'md',
      autoResize = false,
      showCharCount = false,
      fullWidth = true,
      className = '',
      maxLength,
      value,
      onChange,
      placeholder,
      disabled,
      rows,
      name,
      id,
    },
    ref
  ) => {
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const inputId = id || `textarea-${Math.random().toString(36).substr(2, 9)}`;
    const hasError = Boolean(error);
    const charCount = typeof value === 'string' ? value.length : 0;

    const adjustHeight = useCallback(() => {
      const textarea = textareaRef.current;
      if (textarea && autoResize) {
        textarea.style.height = 'auto';
        textarea.style.height = `${textarea.scrollHeight}px`;
      }
    }, [autoResize]);

    useEffect(() => {
      adjustHeight();
    }, [value, adjustHeight]);

    return (
      <div className={`${fullWidth ? 'w-full' : ''} ${className}`}>
        {label && (
          <label
            htmlFor={inputId}
            className="block text-sm font-medium text-neutral-700 dark:text-neutral-300 mb-1.5"
          >
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          id={inputId}
          name={name}
          value={value}
          maxLength={maxLength}
          onChange={onChange}
          placeholder={placeholder}
          disabled={disabled}
          rows={rows}
          aria-invalid={hasError}
          aria-describedby={
            hasError ? `${inputId}-error` : hint ? `${inputId}-hint` : undefined
          }
          className={`
            w-full rounded-lg border transition-colors duration-200
            bg-white dark:bg-neutral-900
            placeholder:text-neutral-400 dark:placeholder:text-neutral-500
            focus:outline-none focus:ring-2 focus:ring-offset-0
            ${sizeClasses[size]}
            ${
              hasError
                ? 'border-error-500 focus:border-error-500 focus:ring-error-500/20'
                : 'border-neutral-300 dark:border-neutral-600 focus:border-primary-500 focus:ring-primary-500/20'
            }
            ${autoResize ? 'resize-none overflow-hidden' : 'resize-y'}
            ${disabled ? 'opacity-50 cursor-not-allowed bg-neutral-50 dark:bg-neutral-800' : ''}
          `}
        />
        <div className="flex items-center justify-between mt-1.5">
          <div>
            {error && (
              <p id={`${inputId}-error`} className="text-sm text-error-500" role="alert">
                {error}
              </p>
            )}
            {hint && !error && (
              <p id={`${inputId}-hint`} className="text-sm text-neutral-500">
                {hint}
              </p>
            )}
          </div>
          {showCharCount && maxLength !== undefined && (
            <span className={`text-xs ${charCount > maxLength ? 'text-error-500' : 'text-neutral-400'}`}>
              {charCount}/{maxLength}
            </span>
          )}
        </div>
      </div>
    );
  }
);

TextArea.displayName = 'TextArea';

export default TextArea;
