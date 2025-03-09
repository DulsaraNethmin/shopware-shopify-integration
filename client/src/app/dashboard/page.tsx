'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import DashboardLayout from '@/components/layout/DashboardLayout';
import PageHeader from '@/components/layout/PageHeader';
import { AuthGuard } from '@/contexts/AuthContext';
import api, { Connector, Dataflow, MigrationLog } from '@/lib/api';
import { formatDate, getStatusVariant } from '@/lib/utils';
import toast from 'react-hot-toast';

export default function DashboardPage() {
    const [isLoading, setIsLoading] = useState(true);
    const [stats, setStats] = useState({
        connectors: 0,
        dataflows: 0,
        successfulMigrations: 0,
        failedMigrations: 0,
    });
    const [recentConnectors, setRecentConnectors] = useState<Connector[]>([]);
    const [recentDataflows, setRecentDataflows] = useState<Dataflow[]>([]);
    const [recentLogs, setRecentLogs] = useState<MigrationLog[]>([]);

    useEffect(() => {
        async function fetchDashboardData() {
            setIsLoading(true);
            try {
                // Fetch connectors
                const connectorsRes = await api.getConnectors();
                const connectors = connectorsRes.data.data || [];
                setRecentConnectors(connectors.slice(0, 5));

                // Fetch dataflows
                const dataflowsRes = await api.getDataflows();
                const dataflows = dataflowsRes.data.data || [];
                setRecentDataflows(dataflows.slice(0, 5));

                // Fetch logs from all dataflows
                let allLogs: MigrationLog[] = [];
                for (const dataflow of dataflows) {
                    try {
                        const logsRes = await api.getMigrationLogs(dataflow.id);
                        allLogs = [...allLogs, ...(logsRes.data.data || [])];
                    } catch (error) {
                        console.error(`Error fetching logs for dataflow ${dataflow.id}:`, error);
                    }
                }

                // Sort logs by created_at and take most recent 5
                allLogs.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
                setRecentLogs(allLogs.slice(0, 5));

                // Calculate stats
                const successfulMigrations = allLogs.filter(log => log.status === 'success').length;
                const failedMigrations = allLogs.filter(log => log.status === 'failed').length;

                setStats({
                    connectors: connectors.length,
                    dataflows: dataflows.length,
                    successfulMigrations,
                    failedMigrations,
                });
            } catch (error) {
                console.error('Error fetching dashboard data:', error);
                toast.error('Failed to load dashboard data');
            } finally {
                setIsLoading(false);
            }
        }

        fetchDashboardData();
    }, []);

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title="Dashboard"
                    description="Overview of your Shopware to Shopify integration"
                />

                {isLoading ? (
                    <div className="flex justify-center items-center h-64">
                        <p>Loading dashboard data...</p>
                    </div>
                ) : (
                    <>
                        {/* Stats Cards */}
                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">
                                        Total Connectors
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-3xl font-bold">{stats.connectors}</div>
                                </CardContent>
                            </Card>

                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">
                                        Active Dataflows
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-3xl font-bold">{stats.dataflows}</div>
                                </CardContent>
                            </Card>

                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">
                                        Successful Migrations
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-3xl font-bold text-green-600">{stats.successfulMigrations}</div>
                                </CardContent>
                            </Card>

                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">
                                        Failed Migrations
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-3xl font-bold text-red-600">{stats.failedMigrations}</div>
                                </CardContent>
                            </Card>
                        </div>

                        {/* Recent Items */}
                        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                            {/* Recent Connectors */}
                            <Card>
                                <CardHeader className="flex flex-row items-center justify-between pb-2">
                                    <CardTitle className="text-lg font-medium">Recent Connectors</CardTitle>
                                    <Button variant="outline" size="sm" asChild>
                                        <Link href="/connectors">View All</Link>
                                    </Button>
                                </CardHeader>
                                <CardContent>
                                    {recentConnectors.length > 0 ? (
                                        <ul className="space-y-4">
                                            {recentConnectors.map((connector) => (
                                                <li key={connector.id} className="flex justify-between items-center">
                                                    <div>
                                                        <p className="font-medium">{connector.name}</p>
                                                        <p className="text-sm text-muted-foreground">{connector.type}</p>
                                                    </div>
                                                    <Badge variant={connector.is_active ? 'success' : 'secondary'}>
                                                        {connector.is_active ? 'Active' : 'Inactive'}
                                                    </Badge>
                                                </li>
                                            ))}
                                        </ul>
                                    ) : (
                                        <p className="text-muted-foreground text-sm">No connectors found</p>
                                    )}
                                </CardContent>
                            </Card>

                            {/* Recent Dataflows */}
                            <Card>
                                <CardHeader className="flex flex-row items-center justify-between pb-2">
                                    <CardTitle className="text-lg font-medium">Recent Dataflows</CardTitle>
                                    <Button variant="outline" size="sm" asChild>
                                        <Link href="/dataflows">View All</Link>
                                    </Button>
                                </CardHeader>
                                <CardContent>
                                    {recentDataflows.length > 0 ? (
                                        <ul className="space-y-4">
                                            {recentDataflows.map((dataflow) => (
                                                <li key={dataflow.id} className="flex justify-between items-center">
                                                    <div>
                                                        <p className="font-medium">{dataflow.name}</p>
                                                        <p className="text-sm text-muted-foreground">{dataflow.type}</p>
                                                    </div>
                                                    <Badge variant={getStatusVariant(dataflow.status)}>
                                                        {dataflow.status}
                                                    </Badge>
                                                </li>
                                            ))}
                                        </ul>
                                    ) : (
                                        <p className="text-muted-foreground text-sm">No dataflows found</p>
                                    )}
                                </CardContent>
                            </Card>

                            {/* Recent Migration Logs */}
                            <Card>
                                <CardHeader className="flex flex-row items-center justify-between pb-2">
                                    <CardTitle className="text-lg font-medium">Recent Migrations</CardTitle>
                                    <Button variant="outline" size="sm" asChild>
                                        <Link href="/logs">View All</Link>
                                    </Button>
                                </CardHeader>
                                <CardContent>
                                    {recentLogs.length > 0 ? (
                                        <ul className="space-y-4">
                                            {recentLogs.map((log) => (
                                                <li key={log.id} className="flex justify-between items-center">
                                                    <div>
                                                        <p className="font-medium">ID: {log.source_identifier}</p>
                                                        <p className="text-sm text-muted-foreground">{formatDate(log.created_at)}</p>
                                                    </div>
                                                    <Badge variant={getStatusVariant(log.status)}>
                                                        {log.status}
                                                    </Badge>
                                                </li>
                                            ))}
                                        </ul>
                                    ) : (
                                        <p className="text-muted-foreground text-sm">No migration logs found</p>
                                    )}
                                </CardContent>
                            </Card>
                        </div>
                    </>
                )}
            </DashboardLayout>
        </AuthGuard>
    );
}