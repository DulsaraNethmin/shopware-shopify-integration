'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent } from '@/components/ui/Card';
import DashboardLayout from '@/components/layout/DashboardLayout';
import PageHeader from '@/components/layout/PageHeader';
import { AuthGuard } from '@/contexts/AuthContext';
import api, { Dataflow } from '@/lib/api';
import { formatDate, getStatusVariant } from '@/lib/utils';
import toast from 'react-hot-toast';

export default function DataflowsPage() {
    const [dataflows, setDataflows] = useState<Dataflow[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    const fetchDataflows = async () => {
        try {
            setIsLoading(true);
            const response = await api.getDataflows();
            setDataflows(response.data.data || []);
        } catch (error) {
            console.error('Error fetching dataflows:', error);
            toast.error('Failed to load dataflows');
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchDataflows();
    }, []);

    const handleDeleteDataflow = async (id: number) => {
        if (!window.confirm('Are you sure you want to delete this dataflow?')) {
            return;
        }

        try {
            await api.deleteDataflow(id);
            toast.success('Dataflow deleted successfully');
            fetchDataflows();
        } catch (error) {
            console.error('Error deleting dataflow:', error);
            toast.error('Failed to delete dataflow');
        }
    };

    const toggleDataflowStatus = async (dataflow: Dataflow) => {
        try {
            const newStatus = dataflow.status === 'active' ? 'inactive' : 'active';
            await api.updateDataflow(dataflow.id, { ...dataflow, status: newStatus });
            toast.success(`Dataflow ${newStatus === 'active' ? 'activated' : 'deactivated'} successfully`);
            fetchDataflows();
        } catch (error) {
            console.error('Error updating dataflow status:', error);
            toast.error('Failed to update dataflow status');
        }
    };

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title="Dataflows"
                    description="Manage your data migration flows"
                    actions={
                        <Button asChild variant={"outline"}>
                            <Link href="/dataflows/new">Create Dataflow</Link>
                        </Button>
                    }
                />

                {isLoading ? (
                    <div className="flex justify-center items-center h-64">
                        <p>Loading dataflows...</p>
                    </div>
                ) : dataflows.length === 0 ? (
                    <Card>
                        <CardContent className="flex flex-col items-center justify-center h-64">
                            <p className="text-muted-foreground mb-4">No dataflows found</p>
                            <Button asChild variant={"secondary"}>
                                <Link href="/dataflows/new">Create Your First Dataflow</Link>
                            </Button>
                        </CardContent>
                    </Card>
                ) : (
                    <div className="grid gap-6">
                        {dataflows.map((dataflow) => (
                            <Card key={dataflow.id} className="overflow-hidden">
                                <div className="p-6">
                                    <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-4">
                                        <div className="mb-4 md:mb-0">
                                            <div className="flex items-center gap-2">
                                                <h3 className="text-lg font-semibold">{dataflow.name}</h3>
                                                <Badge variant={getStatusVariant(dataflow.status)}>
                                                    {dataflow.status}
                                                </Badge>
                                                <Badge variant="secondary">{dataflow.type}</Badge>
                                            </div>
                                            <p className="text-sm text-muted-foreground">{dataflow.description}</p>
                                        </div>
                                        <div className="flex flex-wrap gap-2">
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => toggleDataflowStatus(dataflow)}
                                            >
                                                {dataflow.status === 'active' ? 'Deactivate' : 'Activate'}
                                            </Button>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                asChild
                                            >
                                                <Link href={`/dataflows/${dataflow.id}`}>Manage</Link>
                                            </Button>
                                            <Button
                                                variant="destructive"
                                                size="sm"
                                                onClick={() => handleDeleteDataflow(dataflow.id)}
                                            >
                                                Delete
                                            </Button>
                                        </div>
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 border-t pt-4">
                                        <div>
                                            <h4 className="text-sm font-medium mb-2">Source Connector</h4>
                                            <div className="flex items-center gap-2">
                                                <Badge variant="info">{dataflow.source_connector.type}</Badge>
                                                <span className="text-sm">{dataflow.source_connector.name}</span>
                                            </div>
                                        </div>
                                        <div>
                                            <h4 className="text-sm font-medium mb-2">Destination Connector</h4>
                                            <div className="flex items-center gap-2">
                                                <Badge variant="info">{dataflow.dest_connector.type}</Badge>
                                                <span className="text-sm">{dataflow.dest_connector.name}</span>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="flex justify-between items-center mt-4 pt-4 border-t text-sm text-muted-foreground">
                                        <div>Created: {formatDate(dataflow.created_at)}</div>
                                        <Link
                                            href={`/dataflows/${dataflow.id}/mappings`}
                                            className="text-blue-600 hover:underline"
                                        >
                                            Manage Field Mappings
                                        </Link>
                                    </div>
                                </div>
                            </Card>
                        ))}
                    </div>
                )}
            </DashboardLayout>
        </AuthGuard>
    );
}