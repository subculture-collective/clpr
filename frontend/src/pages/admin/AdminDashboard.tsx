import { Link, useNavigate } from 'react-router-dom';
import { Container, Grid, Card, CardHeader, CardBody } from '../../components';

const quickDocLinks = [
    {
        name: 'API Reference',
        path: '/admin/api-docs',
        description: 'Interactive API documentation',
        isRoute: true,
    },
    {
        name: 'Runbook',
        path: 'operations/runbook',
        description: 'Incident response procedures',
    },
    {
        name: 'Deployment',
        path: 'operations/deployment',
        description: 'Deploy to production',
    },
    {
        name: 'Monitoring',
        path: 'operations/monitoring',
        description: 'Metrics and alerts',
    },
    {
        name: 'Database',
        path: 'backend/database',
        description: 'Schema and migrations',
    },
    {
        name: 'Feature Flags',
        path: 'operations/feature-flags',
        description: 'Toggle features',
    },
];

export function AdminDashboard() {
    const navigate = useNavigate();

    const handleDocClick = (path: string, isRoute: boolean = false) => {
        if (isRoute) {
            // Navigate to route directly
            navigate(path);
        } else {
            // Navigate to docs page with the specific document
            navigate(`/docs?doc=${path}`);
        }
    };

    return (
        <Container className='py-4 xs:py-6 md:py-8'>
            <h1 className='text-2xl xs:text-3xl font-bold mb-6 xs:mb-8'>
                Admin Dashboard
            </h1>

            {/* Quick Documentation Access */}
            <Card className='mb-6 xs:mb-8'>
                <CardHeader>
                    <div className='flex justify-between items-center'>
                        <h2 className='text-xl font-semibold'>
                            📚 Quick Documentation
                        </h2>
                        <Link
                            to='/docs'
                            className='text-sm text-primary hover:underline'
                        >
                            View All Docs →
                        </Link>
                    </div>
                </CardHeader>
                <CardBody>
                    <Grid cols={1} gap={3} responsive={{ sm: 2, md: 3 }}>
                        {quickDocLinks.map(doc => (
                            <button
                                key={doc.path}
                                onClick={() => handleDocClick(doc.path, doc.isRoute)}
                                className='text-left p-3 border border-border rounded-lg hover:bg-accent transition-colors'
                            >
                                <h3 className='font-semibold text-sm mb-1'>
                                    {doc.name}
                                </h3>
                                <p className='text-xs text-muted-foreground'>
                                    {doc.description}
                                </p>
                            </button>
                        ))}
                    </Grid>
                </CardBody>
            </Card>

            {/* Admin Tools */}
            <h2 className='text-xl font-semibold mb-4'>Admin Tools</h2>
            <Grid
                cols={1}
                gap={4}
                responsive={{ sm: 1, md: 2, lg: 3 }}
                className='xs:gap-6'
            >
                <Link to='/admin/clips' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Clip Moderation
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Review and moderate clips submitted to the
                                platform
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/comments' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Comment Moderation
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Manage and moderate user comments
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/users' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                User Management
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Manage user accounts and permissions
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/reports' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Reports
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Review user reports and take action
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/sync' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Sync Controls
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Manually trigger Twitch clip synchronization
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/analytics' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Platform Analytics
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                View platform metrics and user engagement
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/revenue' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Revenue Dashboard
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                MRR, churn, ARPU, and subscription metrics
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/verification' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Creator Verification
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Review and manage creator verification
                                applications
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/campaigns' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Ad Campaigns
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Manage campaigns, creatives, and view
                                performance
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/discovery-lists' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Discovery Lists
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Create and manage curated discovery lists
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/webhooks/dlq' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Webhook DLQ
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                View and replay failed webhook deliveries
                            </p>
                        </CardBody>
                    </Card>
                </Link>

                <Link to='/admin/tags' className='touch-target'>
                    <Card hover clickable>
                        <CardHeader>
                            <h3 className='text-lg xs:text-xl font-semibold'>
                                Tag Management
                            </h3>
                        </CardHeader>
                        <CardBody>
                            <p className='text-sm xs:text-base text-muted-foreground'>
                                Manage blacklisted tag patterns and filters
                            </p>
                        </CardBody>
                    </Card>
                </Link>

            </Grid>
        </Container>
    );
}
