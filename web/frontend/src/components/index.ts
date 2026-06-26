// UI Components
export { Button } from './ui/Button';
export type { ButtonProps } from './ui/Button';
export { Badge } from './ui/Badge';
export type { BadgeProps, DifficultyLevel, BadgeSize } from './ui/Badge';
type BadgeStatus = 'pending' | 'running' | 'accepted' | 'wrong_answer' | 'timeout' | 'compilation_error';
export { Card, CardHeader, CardTitle, CardContent, CardFooter } from './ui/Card';
export type { CardProps } from './ui/Card';
export { Modal } from './ui/Modal';
export type { ModalProps } from './ui/Modal';
export { Spinner } from './ui/Spinner';
export type { SpinnerProps } from './ui/Spinner';
export { Toast, ToastContainer } from './ui/Toast';
export type { ToastProps, ToastContainerProps, ToastVariant, ToastPosition } from './ui/Toast';
export { Input } from './ui/Input';
export type { InputProps } from './ui/Input';
export { Select } from './ui/Select';
export type { SelectProps, SelectOption } from './ui/Select';
export { TextArea } from './ui/TextArea';
export type { TextAreaProps } from './ui/TextArea';

// Layout Components
export { Header } from './layout/Header';
export type { HeaderProps, NavLink, User } from './layout/Header';
export { Sidebar } from './layout/Sidebar';
export type { SidebarProps, SidebarCategory } from './layout/Sidebar';
export { Footer } from './layout/Footer';
export type { FooterProps, FooterSection, SocialLink } from './layout/Footer';

// Problem Components
export { ProblemCard } from './problem/ProblemCard';
export type { ProblemCardProps, ProblemDifficulty, ProblemStatus } from './problem/ProblemCard';
export { ProblemFilters } from './problem/ProblemFilters';
export type { ProblemFiltersProps } from './problem/ProblemFilters';
export { ProblemDescription } from './problem/ProblemDescription';
export type { ProblemDescriptionProps, ExampleBlock, Constraint } from './problem/ProblemDescription';
export { CategoryIcon, CATEGORY_LIST, getCategoryColor, getCategoryLabel } from './problem/CategoryIcon';
export type { } from './problem/CategoryIcon';

// Submission Components
export { SubmissionStatus } from './submission/SubmissionStatus';
export type { SubmissionStatusProps } from './submission/SubmissionStatus';
export { TestCaseResult } from './submission/TestCaseResult';
export type { TestCaseResultProps, TestCase } from './submission/TestCaseResult';

// Performance Components
export { LoadingSpinner, LoadingShell, PageLoader } from './ui/LoadingSpinner';
export { LazyImage } from './ui/LazyImage';
export { PWAInstallPrompt, PWAUpdateBanner, OfflineIndicator } from './PWAComponents';

// Styles
export { colors, typography, spacing, borderRadius, shadows, animation, breakpoints, zIndex } from '../styles/tokens';
