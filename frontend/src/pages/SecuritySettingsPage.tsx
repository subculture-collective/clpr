import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { Alert, Button, Card, CardBody, CardHeader, Container, Input, Spinner, Stack } from '../components';
import {
    disableMFA,
    getMFAStatus,
    getTrustedDevices,
    mfaErrorMessage,
    regenerateBackupCodes,
    revokeTrustedDevice,
    startMFAEnrollment,
    verifyMFAEnrollment,
    type MFAEnrollment,
} from '../lib/mfa-api';

const statusKey = ['mfa', 'status'] as const;
const devicesKey = ['mfa', 'trusted-devices'] as const;

const isSixDigitCode = (value: string) => /^\d{6}$/.test(value.replace(/\s/g, ''));

function formatDate(value?: string) {
    if (!value) return '';
    return new Date(value).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}

function BackupCodeList({ codes }: { codes: string[] }) {
    return (
        <div className='space-y-2'>
            <ul className='grid grid-cols-2 gap-2 rounded-lg border border-border bg-muted/40 p-4 font-mono text-sm' aria-label='Backup codes'>
                {codes.map(code => (
                    <li key={code}>{code}</li>
                ))}
            </ul>
            <p className='text-sm text-muted-foreground'>
                Store these somewhere safe. Each code signs you in once if you lose your authenticator. They will not be shown again.
            </p>
        </div>
    );
}

function CodeForm({
    label,
    submitLabel,
    variant = 'primary',
    pending,
    onSubmit,
}: {
    label: string;
    submitLabel: string;
    variant?: 'primary' | 'danger';
    pending: boolean;
    onSubmit: (code: string) => void;
}) {
    const [code, setCode] = useState('');
    const handleSubmit = (event: FormEvent) => {
        event.preventDefault();
        if (isSixDigitCode(code)) onSubmit(code);
    };
    return (
        <form onSubmit={handleSubmit} className='flex flex-col gap-3 sm:flex-row sm:items-end' noValidate>
            <div className='sm:w-48'>
                <Input
                    label={label}
                    value={code}
                    onChange={event => setCode(event.target.value)}
                    inputMode='numeric'
                    autoComplete='one-time-code'
                    maxLength={7}
                    placeholder='123456'
                    disabled={pending}
                />
            </div>
            <Button type='submit' variant={variant} loading={pending} disabled={!isSixDigitCode(code)}>
                {submitLabel}
            </Button>
        </form>
    );
}

function EnrollmentSetup({ onDone }: { onDone: () => void }) {
    const queryClient = useQueryClient();
    const [enrollment, setEnrollment] = useState<MFAEnrollment | null>(null);

    const start = useMutation({ mutationFn: () => startMFAEnrollment(), onSuccess: setEnrollment });
    const verify = useMutation({
        mutationFn: (code: string) => verifyMFAEnrollment(code),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: statusKey });
            onDone();
        },
    });

    if (!enrollment) {
        return (
            <Stack direction='vertical' gap={3}>
                <p className='text-muted-foreground'>
                    Two-factor authentication asks for a code from an authenticator app (such as 1Password, Authy, or Google Authenticator) each time you sign in on a new browser.
                </p>
                {start.isError && <Alert variant='error'>{mfaErrorMessage(start.error, 'Could not start setup. Please try again.')}</Alert>}
                <div>
                    <Button variant='primary' loading={start.isPending} onClick={() => start.mutate()}>
                        Set up two-factor authentication
                    </Button>
                </div>
            </Stack>
        );
    }

    return (
        <Stack direction='vertical' gap={4}>
            <div>
                <h3 className='font-semibold'>1. Scan this QR code</h3>
                <p className='text-sm text-muted-foreground'>Scan it with your authenticator app, or enter the setup key by hand.</p>
                <img src={enrollment.qr_code_url} alt='QR code for your authenticator app' width={192} height={192} className='mt-3 rounded-lg bg-white p-2' />
                <p className='mt-2 text-sm'>
                    Setup key: <code className='break-all font-mono'>{enrollment.secret}</code>
                </p>
            </div>
            <div>
                <h3 className='font-semibold'>2. Save your backup codes</h3>
                <BackupCodeList codes={enrollment.backup_codes} />
            </div>
            <div>
                <h3 className='mb-2 font-semibold'>3. Enter the code from your app</h3>
                {verify.isError && <Alert variant='error' className='mb-3'>{mfaErrorMessage(verify.error, 'That code is incorrect. Try again.')}</Alert>}
                <CodeForm label='Authentication code' submitLabel='Turn on' pending={verify.isPending} onSubmit={code => verify.mutate(code)} />
            </div>
        </Stack>
    );
}

function TrustedDevices() {
    const queryClient = useQueryClient();
    const devices = useQuery({ queryKey: devicesKey, queryFn: getTrustedDevices });
    const revoke = useMutation({
        mutationFn: (id: number) => revokeTrustedDevice(id),
        onSettled: () => {
            void queryClient.invalidateQueries({ queryKey: devicesKey });
            void queryClient.invalidateQueries({ queryKey: statusKey });
        },
    });

    if (devices.isLoading) return <Spinner size='sm' />;
    if (devices.isError) return <Alert variant='error'>Could not load trusted browsers.</Alert>;
    if (!devices.data?.length) {
        return <p className='text-sm text-muted-foreground'>No trusted browsers. You will be asked for a code on every sign-in.</p>;
    }
    return (
        <ul className='divide-y divide-border rounded-lg border border-border'>
            {devices.data.map(device => (
                <li key={device.id} className='flex flex-col gap-2 p-3 sm:flex-row sm:items-center sm:justify-between'>
                    <div className='min-w-0'>
                        <p className='truncate text-sm font-medium' title={device.device_name}>{device.device_name || 'Unknown browser'}</p>
                        <p className='text-xs text-muted-foreground'>
                            Trusted {formatDate(device.trusted_at)} · expires {formatDate(device.expires_at)}
                        </p>
                    </div>
                    <Button
                        variant='outline'
                        size='sm'
                        loading={revoke.isPending && revoke.variables === device.id}
                        onClick={() => revoke.mutate(device.id)}
                    >
                        Revoke
                    </Button>
                </li>
            ))}
        </ul>
    );
}

function EnabledSettings() {
    const queryClient = useQueryClient();
    const status = useQuery({ queryKey: statusKey, queryFn: getMFAStatus });
    const [newCodes, setNewCodes] = useState<string[] | null>(null);

    const regenerate = useMutation({
        mutationFn: (code: string) => regenerateBackupCodes(code),
        onSuccess: codes => {
            setNewCodes(codes);
            void queryClient.invalidateQueries({ queryKey: statusKey });
        },
    });
    const disable = useMutation({
        mutationFn: (code: string) => disableMFA(code),
        onSuccess: async () => {
            queryClient.removeQueries({ queryKey: devicesKey });
            await queryClient.invalidateQueries({ queryKey: statusKey });
        },
    });

    const data = status.data;
    return (
        <Stack direction='vertical' gap={6}>
            <Alert variant='success'>
                Two-factor authentication is on{data?.enrolled_at ? ` since ${formatDate(data.enrolled_at)}` : ''}.
            </Alert>

            <section aria-labelledby='backup-codes-heading' className='space-y-3'>
                <h3 id='backup-codes-heading' className='font-semibold'>Backup codes</h3>
                <p className='text-sm text-muted-foreground'>
                    {data?.backup_codes_remaining ?? 0} unused backup codes left. Generating new codes replaces all old ones.
                </p>
                {newCodes ?
                    <BackupCodeList codes={newCodes} />
                :   <>
                        {regenerate.isError && <Alert variant='error'>{mfaErrorMessage(regenerate.error, 'That code is incorrect. Try again.')}</Alert>}
                        <CodeForm label='Authentication code' submitLabel='Generate new codes' pending={regenerate.isPending} onSubmit={code => regenerate.mutate(code)} />
                    </>
                }
            </section>

            <section aria-labelledby='trusted-browsers-heading' className='space-y-3'>
                <h3 id='trusted-browsers-heading' className='font-semibold'>Trusted browsers</h3>
                <TrustedDevices />
            </section>

            <section aria-labelledby='disable-mfa-heading' className='space-y-3'>
                <h3 id='disable-mfa-heading' className='font-semibold'>Turn off two-factor authentication</h3>
                {data?.required && (
                    <Alert variant='warning'>Your role requires two-factor authentication. Turning it off blocks moderation and admin actions.</Alert>
                )}
                {disable.isError && <Alert variant='error'>{mfaErrorMessage(disable.error, 'That code is incorrect. Try again.')}</Alert>}
                <CodeForm label='Authentication code' submitLabel='Turn off' variant='danger' pending={disable.isPending} onSubmit={code => disable.mutate(code)} />
            </section>
        </Stack>
    );
}

export function SecuritySettingsPage() {
    const status = useQuery({ queryKey: statusKey, queryFn: getMFAStatus });
    const [justEnabled, setJustEnabled] = useState(false);

    return (
        <Container className='py-4 xs:py-6 md:py-8'>
            <div className='mx-auto max-w-4xl'>
                <p className='mb-2 text-sm'>
                    <Link to='/settings' className='text-muted-foreground hover:underline'>Settings</Link>
                </p>
                <h1 className='mb-6 text-2xl font-bold xs:text-3xl'>Security</h1>
                <Card>
                    <CardHeader>
                        <h2 className='text-xl font-semibold'>Two-factor authentication</h2>
                    </CardHeader>
                    <CardBody>
                        {status.isLoading && <Spinner size='md' />}
                        {status.isError && <Alert variant='error'>Could not load your two-factor authentication status.</Alert>}
                        {status.data && !status.data.enabled && (
                            <Stack direction='vertical' gap={4}>
                                {status.data.required && (
                                    <Alert variant='warning'>
                                        Your role requires two-factor authentication
                                        {status.data.in_grace_period && status.data.grace_period_end ?
                                            `. Set it up by ${formatDate(status.data.grace_period_end)} to keep using moderation and admin tools.`
                                        :   '. Moderation and admin actions are blocked until you set it up.'}
                                    </Alert>
                                )}
                                <EnrollmentSetup onDone={() => setJustEnabled(true)} />
                            </Stack>
                        )}
                        {status.data?.enabled && (
                            <Stack direction='vertical' gap={4}>
                                {justEnabled && (
                                    <Alert variant='info'>From now on, signing in on a new browser asks for a code from your authenticator app.</Alert>
                                )}
                                <EnabledSettings />
                            </Stack>
                        )}
                    </CardBody>
                </Card>
            </div>
        </Container>
    );
}
