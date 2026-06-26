import React, { HTMLAttributes, forwardRef } from 'react';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'problem' | 'submission' | 'stats' | 'default';
  shadow?: 'none' | 'sm' | 'md' | 'lg' | 'xl' | 'glow';
  hoverable?: boolean;
  padding?: 'none' | 'sm' | 'md' | 'lg';
}

const shadowClasses: Record<string, string> = {
  none: 'shadow-none',
  sm: 'shadow-sm',
  md: 'shadow-md',
  lg: 'shadow-lg',
  xl: 'shadow-xl',
  glow: 'shadow-glow',
};

const paddingClasses: Record<string, string> = {
  none: '',
  sm: 'p-3',
  md: 'p-4',
  lg: 'p-6',
};

const variantClasses: Record<string, string> = {
  default: 'bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-700',
  problem: 'bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-700',
  submission: 'bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-700',
  stats: 'bg-gradient-to-br from-primary-50 to-secondary-50 dark:from-primary-950 dark:to-secondary-950 border border-primary-100 dark:border-primary-900',
};

const hoverClasses: Record<string, string> = {
  problem: 'hover:shadow-lg hover:-translate-y-1 hover:border-primary-300 dark:hover:border-primary-700',
  submission: 'hover:shadow-md hover:border-neutral-300 dark:hover:border-neutral-600',
  stats: 'hover:shadow-lg hover:scale-[1.02]',
  default: 'hover:shadow-md',
};

export const Card = forwardRef<HTMLDivElement, CardProps>(
  (
    {
      variant = 'default',
      shadow = 'md',
      hoverable = true,
      padding = 'md',
      children,
      className = '',
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={`
          rounded-xl
          ${variantClasses[variant]}
          ${shadowClasses[shadow]}
          ${paddingClasses[padding]}
          ${hoverable ? hoverClasses[variant] : ''}
          transition-all duration-300 ease-in-out
          ${className}
        `}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Card.displayName = 'Card';

export interface CardHeaderProps extends HTMLAttributes<HTMLDivElement> {}

export const CardHeader: React.FC<CardHeaderProps> = ({ children, className = '', ...props }) => (
  <div className={`mb-4 ${className}`} {...props}>
    {children}
  </div>
);

export interface CardTitleProps extends HTMLAttributes<HTMLHeadingElement> {}

export const CardTitle: React.FC<CardTitleProps> = ({ children, className = '', ...props }) => (
  <h3 className={`text-lg font-semibold text-neutral-900 dark:text-neutral-100 ${className}`} {...props}>
    {children}
  </h3>
);

export interface CardContentProps extends HTMLAttributes<HTMLDivElement> {}

export const CardContent: React.FC<CardContentProps> = ({ children, className = '', ...props }) => (
  <div className={`text-neutral-600 dark:text-neutral-400 ${className}`} {...props}>
    {children}
  </div>
);

export interface CardFooterProps extends HTMLAttributes<HTMLDivElement> {}

export const CardFooter: React.FC<CardFooterProps> = ({ children, className = '', ...props }) => (
  <div className={`mt-4 pt-4 border-t border-neutral-200 dark:border-neutral-700 ${className}`} {...props}>
    {children}
  </div>
);

export default Card;
