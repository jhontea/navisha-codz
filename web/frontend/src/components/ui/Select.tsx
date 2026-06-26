import React, { useState, useRef, useEffect } from 'react';

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

export interface SelectProps {
  options: SelectOption[];
  value?: string | string[];
  onChange?: (value: string | string[]) => void;
  placeholder?: string;
  label?: string;
  error?: string;
  hint?: string;
  size?: 'sm' | 'md' | 'lg';
  searchable?: boolean;
  multiSelect?: boolean;
  clearable?: boolean;
  disabled?: boolean;
  fullWidth?: boolean;
  className?: string;
}

export const Select: React.FC<SelectProps> = ({
  options,
  value,
  onChange,
  placeholder = 'Select an option',
  label,
  error,
  hint,
  size = 'md',
  searchable = false,
  multiSelect = false,
  clearable = true,
  disabled = false,
  fullWidth = true,
  className = '',
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
        setSearchTerm('');
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Focus input when dropdown opens
  useEffect(() => {
    if (isOpen && searchable && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen, searchable]);

  const selectedValues = multiSelect
    ? Array.isArray(value) ? value : []
    : value ? [value] : [];

  const filteredOptions = searchTerm
    ? options.filter((opt) =>
        opt.label.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : options;

  const getDisplayLabel = () => {
    if (multiSelect) {
      if (selectedValues.length === 0) return placeholder;
      if (selectedValues.length === 1) {
        return options.find((o) => o.value === selectedValues[0])?.label || '';
      }
      return `${selectedValues.length} selected`;
    }
    if (!value) return placeholder;
    return options.find((o) => o.value === value)?.label || placeholder;
  };

  const handleSelect = (optionValue: string) => {
    if (multiSelect) {
      const current = Array.isArray(value) ? value : [];
      const updated = current.includes(optionValue)
        ? current.filter((v) => v !== optionValue)
        : [...current, optionValue];
      onChange?.(updated);
    } else {
      onChange?.(optionValue);
      setIsOpen(false);
      setSearchTerm('');
    }
  };

  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation();
    onChange?.(multiSelect ? [] : '');
  };

  const sizeClasses: Record<string, string> = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-3.5 py-2 text-base',
    lg: 'px-4 py-3 text-lg',
  };

  const hasError = Boolean(error);

  return (
    <div className={`${fullWidth ? 'w-full' : ''} ${className}`} ref={containerRef}>
      {label && (
        <label className="block text-sm font-medium text-neutral-700 dark:text-neutral-300 mb-1.5">
          {label}
        </label>
      )}
      <div className="relative">
        {/* Trigger */}
        <button
          type="button"
          onClick={() => !disabled && setIsOpen(!isOpen)}
          aria-haspopup="listbox"
          aria-expanded={isOpen}
          disabled={disabled}
          className={`
            w-full ${sizeClasses[size]} rounded-lg border text-left
            bg-white dark:bg-neutral-900
            transition-colors duration-200
            flex items-center justify-between
            ${isOpen ? 'ring-2' : ''}
            ${
              hasError
                ? 'border-error-500 ring-error-500/20'
                : isOpen
                ? 'border-primary-500 ring-primary-500/20'
                : 'border-neutral-300 dark:border-neutral-600'
            }
            ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
          `}
        >
          <span className={`truncate ${!value || (Array.isArray(value) && value.length === 0) ? 'text-neutral-400' : 'text-neutral-900 dark:text-neutral-100'}`}>
            {getDisplayLabel()}
          </span>
          <span className="flex items-center gap-1">
            {clearable && selectedValues.length > 0 && (
              <span
                onClick={handleClear}
                className="p-0.5 rounded hover:bg-neutral-100 dark:hover:bg-neutral-800"
                role="button"
                aria-label="Clear selection"
              >
                <svg className="w-4 h-4 text-neutral-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </span>
            )}
            <svg className={`w-4 h-4 text-neutral-400 transition-transform ${isOpen ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </span>
        </button>

        {/* Dropdown */}
        {isOpen && (
          <div
            className="absolute z-dropdown w-full mt-1 bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-700 rounded-lg shadow-lg overflow-hidden"
            role="listbox"
          >
            {/* Search Input */}
            {searchable && (
              <div className="p-2 border-b border-neutral-200 dark:border-neutral-700">
                <input
                  ref={inputRef}
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Search..."
                  className="w-full px-3 py-1.5 text-sm rounded border border-neutral-300 dark:border-neutral-600 bg-white dark:bg-neutral-800 focus:outline-none focus:ring-1 focus:ring-primary-500"
                />
              </div>
            )}

            {/* Options */}
            <div className="max-h-60 overflow-y-auto py-1">
              {filteredOptions.length === 0 ? (
                <div className="px-3 py-2 text-sm text-neutral-500">No options found</div>
              ) : (
                filteredOptions.map((option) => {
                  const isSelected = selectedValues.includes(option.value);
                  return (
                    <button
                      key={option.value}
                      type="button"
                      onClick={() => !option.disabled && handleSelect(option.value)}
                      role="option"
                      aria-selected={isSelected}
                      disabled={option.disabled}
                      className={`
                        w-full px-3 py-2 text-left text-sm flex items-center gap-2
                        transition-colors
                        ${isSelected ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300' : 'text-neutral-700 dark:text-neutral-300 hover:bg-neutral-50 dark:hover:bg-neutral-800'}
                        ${option.disabled ? 'opacity-50 cursor-not-allowed' : ''}
                      `}
                    >
                      {multiSelect && (
                        <span className={`w-4 h-4 rounded border flex items-center justify-center ${isSelected ? 'bg-primary-500 border-primary-500' : 'border-neutral-300'}`}>
                          {isSelected && (
                            <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                            </svg>
                          )}
                        </span>
                      )}
                      {option.label}
                    </button>
                  );
                })
              )}
            </div>
          </div>
        )}
      </div>
      {error && (
        <p className="mt-1.5 text-sm text-error-500" role="alert">{error}</p>
      )}
      {hint && !error && (
        <p className="mt-1.5 text-sm text-neutral-500">{hint}</p>
      )}
    </div>
  );
};

Select.displayName = 'Select';

export default Select;
