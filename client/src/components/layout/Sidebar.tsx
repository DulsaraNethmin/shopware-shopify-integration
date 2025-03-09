'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';

interface NavItem {
    href: string;
    title: string;
    icon: React.ReactNode;
}

export default function Sidebar() {
    const pathname = usePathname();

    const navItems: NavItem[] = [
        {
            href: '/dashboard',
            title: 'Dashboard',
            icon: <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path d="M2 10a8 8 0 018-8v8h8a8 8 0 11-16 0z"></path><path d="M12 2.252A8.014 8.014 0 0117.748 8H12V2.252z"></path></svg>
        },
        {
            href: '/connectors',
            title: 'Connectors',
            icon: <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fillRule="evenodd" d="M5 4a3 3 0 00-3 3v6a3 3 0 003 3h10a3 3 0 003-3V7a3 3 0 00-3-3H5zm-1 9v-1h5v2H5a1 1 0 01-1-1zm7 1h4a1 1 0 001-1v-1h-5v2zm0-4h5V8h-5v2zM9 8H4v2h5V8z" clipRule="evenodd"></path></svg>
        },
        {
            href: '/dataflows',
            title: 'Dataflows',
            icon: <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"></path><path fillRule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clipRule="evenodd"></path></svg>
        },
        {
            href: '/logs',
            title: 'Migration Logs',
            icon: <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path d="M2 5a2 2 0 012-2h7a2 2 0 012 2v4a2 2 0 01-2 2H9l-3 3v-3H4a2 2 0 01-2-2V5z"></path><path d="M15 7v2a4 4 0 01-4 4H9.828l-1.766 1.767c.28.149.599.233.938.233h2l3 3v-3h2a2 2 0 002-2V9a2 2 0 00-2-2h-1z"></path></svg>
        }
    ];

    return (
        <aside className="fixed top-0 left-0 z-20 w-64 h-full pt-16 flex flex-col flex-shrink-0 transition-width duration-75 bg-white border-r border-gray-200 overflow-y-auto">
            <div className="relative flex-1 flex flex-col pt-5 pb-4">
                <div className="flex-1 px-3 bg-white divide-y space-y-1">
                    <ul className="space-y-2 pb-2">
                        {navItems.map((item) => (
                            <li key={item.href}>
                                <Link
                                    href={item.href}
                                    className={cn(
                                        "flex items-center p-2 text-base font-normal rounded-lg",
                                        pathname === item.href || pathname.startsWith(`${item.href}/`)
                                            ? "text-white bg-blue-600 hover:bg-blue-700"
                                            : "text-gray-900 hover:bg-gray-100"
                                    )}
                                >
                                    <span className="mr-3">{item.icon}</span>
                                    <span className="ml-3">{item.title}</span>
                                </Link>
                            </li>
                        ))}
                    </ul>
                </div>
            </div>
        </aside>
    );
}