interface PageHeaderProps {
    title: string;
    description?: string;
    actions?: React.ReactNode;
}

export default function PageHeader({ title, description, actions }: PageHeaderProps) {
    return (
        <div className="flex justify-between items-start mb-6">
            <div>
                <h1 className="text-2xl font-semibold text-gray-900">{title}</h1>
                {description && <p className="mt-1 text-sm text-gray-500">{description}</p>}
            </div>
            {actions && <div>{actions}</div>}
        </div>
    );
}