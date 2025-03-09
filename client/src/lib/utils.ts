import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { format } from 'date-fns';

// Combine class names with Tailwind
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

// Format dates consistently
export function formatDate(date: string | Date) {
    const d = typeof date === 'string' ? new Date(date) : date;
    return format(d, 'MMM dd, yyyy HH:mm');
}

// Format error messages from API
export function formatError(error: any): string {
    if (typeof error === 'string') {
        return error;
    }

    if (error?.response?.data?.error) {
        return error.response.data.error;
    }

    if (error?.message) {
        return error.message;
    }

    return 'An unknown error occurred';
}

// Get status badge variant based on status
export function getStatusVariant(status: string): 'success' | 'warning' | 'error' | 'info' | 'secondary' {
    switch (status.toLowerCase()) {
        case 'active':
        case 'success':
            return 'success';
        case 'inactive':
            return 'secondary';
        case 'pending':
        case 'in_progress':
            return 'warning';
        case 'failed':
            return 'error';
        default:
            return 'info';
    }
}

// Truncate text with ellipsis
export function truncate(text: string, length: number) {
    if (text.length <= length) {
        return text;
    }
    return text.substring(0, length) + '...';
}

// Format JSON for display
export function formatJSON(json: string) {
    try {
        return JSON.stringify(JSON.parse(json), null, 2);
    } catch (e) {
        return json;
    }
}

// Convert camelCase to Title Case
export function titleCase(text: string) {
    // Add space before capital letters and uppercase the first letter
    const result = text.replace(/([A-Z])/g, ' $1');
    return result.charAt(0).toUpperCase() + result.slice(1);
}