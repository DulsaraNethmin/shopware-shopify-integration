'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Select } from '@/components/ui/Select';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';
import DashboardLayout from '@/components/layout/DashboardLayout';
import PageHeader from '@/components/layout/PageHeader';
import { AuthGuard } from '@/contexts/AuthContext';
import api, { Dataflow, MigrationLog } from '@/lib/api';
import { formatDate, formatJSON, getStatusVariant } from '@/lib/utils';
import toast from 'react-hot-toast';

export default function LogsPage() {
    const [isLoading, setIsLoading] = useState(true);
    const [dataflows, setDataflows] = useState<Dataflow[]>([]);
    const [selectedDataflow, setSelectedDataflow] = useState<number | null>(null);
    const [logs, setLogs] = useState<MigrationLog[]>([]);
    const [selectedLog, setSelectedLog] = useState<MigrationLog | null>(null);

    // Fetch dataflows on mount
    useEffect(() => {
        async function fetchDataflows() {
            try {
                const response = await api.getDataflows();
                const data = response.data.data || [];
                setDataflows(data);

                // Set first dataflow as selected if available
                if (data.length > 0) {
                    setSelectedDataflow(data[0].id);
                }
            } catch (error) {
                console.error('Error fetching dataflows:', error);
                toast.error('Failed to load dataflows');
            } finally {
                setIsLoading(false);
            }
        }

        fetchDataflows();
    }, []);

    // Fetch logs when dataflow is selected
    useEffect(() => {
        async function fetchLogs() {
            if (!selectedDataflow) return;

            setIsLoading(true);
            try {
                const response = await api.getMigrationLogs(selectedDataflow);
                setLogs(response.data.data || []);
            } catch (error) {
                console.error('Error fetching logs:', error);
                toast.error('Failed to load migration logs');
            } finally {
                setIsLoading(false);
            }
        }

        fetchLogs();
    }, [selectedDataflow]);

    const handleViewLog = (log: MigrationLog) => {
        setSelectedLog(log);
    };

    const handleCloseLogDetails = () => {
        setSelectedLog(null);
    };

    const handleDataflowChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        const value = e.target.value;
        setSelectedDataflow(value ? parseInt(value) : null);
        setSelectedLog(null);
    };

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title="Migration Logs"
                    description="View logs of data migrations between systems"
                />

                <Card className="mb-6">
                    <CardContent className="p-4">
                        <div className="flex items-center space-x-4">
                            <div className="w-64">
                                <label htmlFor="dataflow" className="text-sm font-medium block mb-1">
                                    Filter by Dataflow
                                </label>
                                <Select
                                    id="dataflow"
                                    value={selectedDataflow?.toString() || ''}
                                    onChange={handleDataflowChange}
                                >
                                    <option value="">All Dataflows</option>
                                    {dataflows.map(dataflow => (
                                        <option key={dataflow.id} value={dataflow.id}>
                                            {dataflow.name} ({dataflow.type})
                                        </option>
                                    ))}
                                </Select>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {isLoading ? (
                    <div className="flex justify-center items-center h-64">
                        <p>Loading migration logs...</p>
                    </div>
                ) : logs.length === 0 ? (
                    <Card>
                        <CardContent className="flex flex-col items-center justify-center h-64">
                            <p className="text-muted-foreground mb-4">No migration logs found</p>
                            {selectedDataflow && (
                                <Button asChild>
                                    <Link href={`/dataflows/${selectedDataflow}`}>
                                        View Dataflow
                                    </Link>
                                </Button>
                            )}
                        </CardContent>
                    </Card>
                ) : selectedLog ? (
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between">
                            <CardTitle>Migration Log Details</CardTitle>
                            <Button variant="outline" size="sm" onClick={handleCloseLogDetails}>
                                Back to Logs
                            </Button>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <h3 className="text-sm font-medium mb-2">Basic Information</h3>
                                    <dl className="grid grid-cols-[120px_1fr] gap-2 text-sm">
                                        <dt className="font-medium">Status:</dt>
                                        <dd>
                                            <Badge variant={getStatusVariant(selectedLog.status)}>
                                                {selectedLog.status}
                                            </Badge>
                                        </dd>

                                        <dt className="font-medium">Created:</dt>
                                        <dd>{formatDate(selectedLog.created_at)}</dd>

                                        {selectedLog.completed_at && (
                                            <>
                                                <dt className="font-medium">Completed:</dt>
                                                <dd>{formatDate(selectedLog.completed_at)}</dd>
                                            </>
                                        )}

                                        <dt className="font-medium">Source ID:</dt>
                                        <dd>{selectedLog.source_identifier}</dd>

                                        {selectedLog.dest_identifier && (
                                            <>
                                                <dt className="font-medium">Destination ID:</dt>
                                                <dd>{selectedLog.dest_identifier}</dd>
                                            </>
                                        )}
                                    </dl>
                                </div>

                                {selectedLog.error_message && (
                                    <div>
                                        <h3 className="text-sm font-medium mb-2">Error Message</h3>
                                        <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
                                            {selectedLog.error_message}
                                        </div>
                                    </div>
                                )}
                            </div>

                            <Tabs defaultValue="source">
                                <TabsList>
                                    <TabsTrigger value="source">Source Payload</TabsTrigger>
                                    {selectedLog.transformed_payload && (
                                        <TabsTrigger value="transformed">Transformed Payload</TabsTrigger>
                                    )}
                                </TabsList>

                                <TabsContent value="source" className="mt-4">
                                    <Card>
                                        <CardContent className="p-4">
                      <pre className="bg-gray-50 p-4 rounded text-sm font-mono overflow-auto max-h-96">
                        {formatJSON(selectedLog.source_payload || '{}')}
                      </pre>
                                        </CardContent>
                                    </Card>
                                </TabsContent>

                                {selectedLog.transformed_payload && (
                                    <TabsContent value="transformed" className="mt-4">
                                        <Card>
                                            <CardContent className="p-4">
                        <pre className="bg-gray-50 p-4 rounded text-sm font-mono overflow-auto max-h-96">
                          {formatJSON(selectedLog.transformed_payload)}
                        </pre>
                                            </CardContent>
                                        </Card>
                                    </TabsContent>
                                )}
                            </Tabs>
                        </CardContent>
                    </Card>
                ) : (
                    <div className="space-y-4">
                        {logs.map((log) => (
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
                                            onClick={() => handleViewLog(log)}
                                        >
                                            View Details
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                )}
            </DashboardLayout>
        </AuthGuard>
    );
}