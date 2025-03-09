import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const buttonVariants = cva(
    'inline-flex items-center justify-center rounded-xl text-base font-semibold transition-all duration-300 ease-in-out',
    {
        variants: {
            variant: {
                default: 'bg-primary-600 text-white hover:bg-primary-700 focus:ring-4 focus:ring-primary-300/50 shadow-md hover:shadow-lg',
                destructive: 'bg-red-600 text-white hover:bg-red-700 focus:ring-4 focus:ring-red-300/50 shadow-md hover:shadow-lg',
                outline: 'border-2 border-neutral-300 bg-transparent text-neutral-700 hover:bg-neutral-100 focus:ring-4 focus:ring-neutral-300/50',
                secondary: 'bg-neutral-200 text-neutral-800 hover:bg-neutral-300 focus:ring-4 focus:ring-neutral-300/50',
                ghost: 'text-neutral-700 hover:bg-neutral-200 focus:ring-4 focus:ring-neutral-300/50',
                link: 'text-primary-600 hover:text-primary-700 underline-offset-4 hover:underline',
            },
            size: {
                default: 'h-12 px-5 py-2.5 gap-2',
                sm: 'h-9 px-3 text-sm rounded-lg',
                lg: 'h-14 px-8 text-lg rounded-xl',
                icon: 'h-12 w-12 p-0 justify-center',
            },
        },
        defaultVariants: {
            variant: 'default',
            size: 'default',
        },
    }
);

export interface ButtonProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement>,
        VariantProps<typeof buttonVariants> {
    asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, variant, size, asChild = false, ...props }, ref) => {
        return (
            <button
                className={cn(
                    buttonVariants({ variant, size }),
                    'disabled:opacity-50 disabled:cursor-not-allowed',
                    'active:scale-[0.98]',
                    'focus:outline-none',
                    className
                )}
                ref={ref}
                {...props}
            />
        );
    }
);
Button.displayName = 'Button';

export { Button, buttonVariants };