'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';
import DashboardLayout from '@/components/layout/DashboardLayout';
import PageHeader from '@/components/layout/PageHeader';
import { AuthGuard } from '@/contexts/AuthContext';
import api, { Dataflow, FieldMapping, MigrationLog } from '@/lib/api';
import { formatDate, getStatusVariant } from '@/lib/utils';
import toast from 'react-hot-toast';

export default function DataflowDetailPage() {
    const params = useParams();
    const dataflowId = parseInt(params.id as string);
    const router = useRouter();

    const [isLoading, setIsLoading] = useState(true);
    const [dataflow, setDataflow] = useState<Dataflow | null>(null);
    const [fieldMappings, setFieldMappings] = useState<FieldMapping[]>([]);
    const [recentLogs, setRecentLogs] = useState<MigrationLog[]>([]);

    useEffect(() => {
        async function fetchData() {
            setIsLoading(true);
            try {
                // Fetch dataflow details
                const dataflowRes = await api.getDataflow(dataflowId);
                setDataflow(dataflowRes.data.data);

                // Fetch field mappings
                const mappingsRes = await api.getFieldMappings(dataflowId);
                setFieldMappings(mappingsRes.data.data || []);

                // Fetch recent logs
                const logsRes = await api.getMigrationLogs(dataflowId);
                setRecentLogs((logsRes.data.data || []).slice(0, 5));
            } catch (error) {
                console.error('Error fetching data:', error);
                toast.error('Failed to load dataflow details');
                router.push('/dataflows');
            } finally {
                setIsLoading(false);
            }
        }

        fetchData();
    }, [dataflowId, router]);

    const toggleDataflowStatus = async () => {
        if (!dataflow) return;

        try {
            const newStatus = dataflow.status === 'active' ? 'inactive' : 'active';
            await api.updateDataflow(dataflow.id, { ...dataflow, status: newStatus });

            // Update local state
            setDataflow({
                ...dataflow,
                status: newStatus
            });

            toast.success(`Dataflow ${newStatus === 'active' ? 'activated' : 'deactivated'} successfully`);
        } catch (error) {
            console.error('Error updating dataflow status:', error);
            toast.error('Failed to update dataflow status');
        }
    };

    if (isLoading) {
        return (
            <AuthGuard>
                <DashboardLayout>
                    <div className="flex justify-center items-center h-64">
                        <p>Loading dataflow details...</p>
                    </div>
                </DashboardLayout>
            </AuthGuard>
        );
    }

    if (!dataflow) {
        return (
            <AuthGuard>
                <DashboardLayout>
                    <div className="flex justify-center items-center h-64">
                        <p>Dataflow not found</p>
                    </div>
                </DashboardLayout>
            </AuthGuard>
        );
    }

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title={dataflow.name}
                    description={dataflow.description || `${dataflow.type} data migration flow`}
                    actions={
                        <div className="space-x-2">
                            <Button
                                variant={dataflow.status === 'active' ? 'destructive' : 'default'}
                                onClick={toggleDataflowStatus}
                            >
                                {dataflow.status === 'active' ? 'Deactivate' : 'Activate'}
                            </Button>
                            <Button asChild>
                                <Link href={`/dataflows/${dataflow.id}/mappings`}>Manage Field Mappings</Link>
                            </Button>
                        </div>
                    }
                />

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">Status</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center gap-2">
                                <Badge variant={getStatusVariant(dataflow.status)} className="text-lg">
                                    {dataflow.status}
                                </Badge>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">Data Type</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-lg font-medium">{dataflow.type}</div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">Created</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-lg font-medium">{formatDate(dataflow.created_at)}</div>
                        </CardContent>
                    </Card>
                </div>

                <Tabs defaultValue="connectors">
                    <TabsList className="mb-4">
                        <TabsTrigger value="connectors">Connectors</TabsTrigger>
                        <TabsTrigger value="mappings">Field Mappings</TabsTrigger>
                        <TabsTrigger value="logs">Recent Logs</TabsTrigger>
                    </TabsList>

                    <TabsContent value="connectors">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <Card>
                                <CardHeader>
                                    <CardTitle className="flex items-center gap-2">
                                        Source Connector
                                        <Badge variant="secondary">{dataflow.source_connector.type}</Badge>
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <dl className="grid grid-cols-[100px_1fr] gap-2">
                                        <dt className="font-medium">Name:</dt>
                                        <dd>{dataflow.source_connector.name}</dd>

                                        <dt className="font-medium">URL:</dt>
                                        <dd className="truncate">{dataflow.source_connector.url}</dd>

                                        <dt className="font-medium">Status:</dt>
                                        <dd>
                                            <Badge variant={dataflow.source_connector.is_active ? 'success' : 'secondary'}>
                                                {dataflow.source_connector.is_active ? 'Active' : 'Inactive'}
                                            </Badge>
                                        </dd>
                                    </dl>

                                    <div className="mt-4">
                                        <Button variant="outline" size="sm" asChild>
                                            <Link href={`/connectors/${dataflow.source_connector_id}`}>
                                                Edit Connector
                                            </Link>
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>

                            <Card>
                                <CardHeader>
                                    <CardTitle className="flex items-center gap-2">
                                        Destination Connector
                                        <Badge variant="secondary">{dataflow.dest_connector.type}</Badge>
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <dl className="grid grid-cols-[100px_1fr] gap-2">
                                        <dt className="font-medium">Name:</dt>
                                        <dd>{dataflow.dest_connector.name}</dd>

                                        <dt className="font-medium">URL:</dt>
                                        <dd className="truncate">{dataflow.dest_connector.url}</dd>

                                        <dt className="font-medium">Status:</dt>
                                        <dd>
                                            <Badge variant={dataflow.dest_connector.is_active ? 'success' : 'secondary'}>
                                                {dataflow.dest_connector.is_active ? 'Active' : 'Inactive'}
                                            </Badge>
                                        </dd>
                                    </dl>

                                    <div className="mt-4">
                                        <Button variant="outline" size="sm" asChild>
                                            <Link href={`/connectors/${dataflow.dest_connector_id}`}>
                                                Edit Connector
                                            </Link>
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>
                        </div>
                    </TabsContent>

                    <TabsContent value="mappings">
                        {fieldMappings.length === 0 ? (
                            <Card>
                                <CardContent className="flex flex-col items-center justify-center py-12">
                                    <p className="text-muted-foreground mb-4">No field mappings defined yet</p>
                                    <Button asChild>
                                        <Link href={`/dataflows/${dataflow.id}/mappings`}>
                                            Configure Field Mappings
                                        </Link>
                                    </Button>
                                </CardContent>
                            </Card>
                        ) : (
                            <div className="space-y-4">
                                <div className="flex justify-between items-center">
                                    <h3 className="text-lg font-medium">Field Mappings</h3>
                                    <Button variant="outline" size="sm" asChild>
                                        <Link href={`/dataflows/${dataflow.id}/mappings`}>
                                            Manage Mappings
                                        </Link>
                                    </Button>
                                </div>

                                <div className="overflow-x-auto">
                                    <table className="w-full border-collapse">
                                        <thead>
                                        <tr className="bg-muted">
                                            <th className="p-2 text-left">Source Field</th>
                                            <th className="p-2 text-left">Destination Field</th>
                                            <th className="p-2 text-left">Required</th>
                                            <th className="p-2 text-left">Transform Type</th>
                                        </tr>
                                        </thead>
                                        <tbody>
                                        {fieldMappings.map((mapping) => (
                                            <tr key={mapping.id} className="border-b">
                                                <td className="p-2">
                                                    <code className="bg-gray-100 px-1 rounded text-sm">{mapping.source_field}</code>
                                                </td>
                                                <td className="p-2">
                                                    <code className="bg-gray-100 px-1 rounded text-sm">{mapping.dest_field}</code>
                                                </td>
                                                <td className="p-2">
                                                    {mapping.is_required ? (
                                                        <Badge variant="error">Required</Badge>
                                                    ) : (
                                                        <span className="text-muted-foreground">Optional</span>
                                                    )}
                                                </td>
                                                <td className="p-2">
                                                    <Badge variant="secondary">{mapping.transform_type}</Badge>
                                                </td>
                                            </tr>
                                        ))}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        )}
                    </TabsContent>

                    <TabsContent value="logs">
                        {recentLogs.length === 0 ? (
                            <Card>
                                <CardContent className="flex flex-col items-center justify-center py-12">
                                    <p className="text-muted-foreground mb-4">No migration logs yet</p>
                                    <Button asChild>
                                        <Link href={`/logs`}>
                                            View All Logs
                                        </Link>
                                    </Button>
                                </CardContent>
                            </Card>
                        ) : (
                            <div className="space-y-4">
                                <div className="flex justify-between items-center">
                                    <h3 className="text-lg font-medium">Recent Migration Logs</h3>
                                    <Button variant="outline" size="sm" asChild>
                                        <Link href={`/logs`}>
                                            View All Logs
                                        </Link>
                                    </Button>
                                </div>

                                <div className="space-y-2">
                                    {recentLogs.map((log) => (
                                        <Card key={log.id} className="hover:shadow-md transition-shadow">
                                            <CardContent className="p-4">
                                                <div className="flex flex-col md:flex-row md:items-center md:justify-between">
                                                    <div className="mb-2 md:mb-0">
                                                        <div className="flex items-center gap-2 mb-1">
                                                            <Badge variant={getStatusVariant(log.status)}>
                                                                {log.status}
                                                            </Badge>
                                                            <span className="text-sm text-muted-foreground">
                                {formatDate(log.created_at)}
                              </span>
                                                        </div>
                                                        <p className="text-sm">
                                                            <span className="font-medium">Source ID:</span> {log.source_identifier}
                                                            {log.dest_identifier && (
                                                                <>
                                                                    <span className="mx-2">→</span>
                                                                    <span className="font-medium">Dest ID:</span> {log.dest_identifier}
                                                                </>
                                                            )}
                                                        </p>
                                                    </div>
                                                    <Button
                                                        variant="outline"
                                                        size="sm"
                                                        asChild
                                                    >
                                                        <Link href={`/logs?dataflow=${dataflow.id}&log=${log.id}`}>
                                                            View Details
                                                        </Link>
                                                    </Button>
                                                </div>
                                            </CardContent>
                                        </Card>
                                    ))}
                                </div>
                            </div>
                        )}
                    </TabsContent>
                </Tabs>
            </DashboardLayout>
        </AuthGuard>
    );
}