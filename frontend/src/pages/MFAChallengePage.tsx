import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Alert, Button, Card, CardBody, CardHeader, Checkbox, Container, Input } from '../components';
import { useAuth } from '../context/AuthContext';
import { completeSignIn } from '../lib/complete-sign-in';
import { MFALoginError, verifyMFALogin } from '../lib/mfa-api';

/**
 * Second sign-in step for accounts with two-factor authentication. The
 * backend keeps the pending sign-in in an HttpOnly cookie; this page only
 * collects the code.
 */
export function MFAChallengePage() {
    const navigate = useNavigate();
    const { refreshUser } = useAuth();
    const [code, setCode] = useState('');
    const [useBackupCode, setUseBackupCode] = useState(false);
    const [trustDevice, setTrustDevice] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [signInExpired, setSignInExpired] = useState(false);

    const trimmed = code.replace(/[\s-]/g, '');
    const canSubmit = useBackupCode ? trimmed.length === 8 : /^\d{6}$/.test(trimmed);

    const handleSubmit = async (event: FormEvent) => {
        event.preventDefault();
        if (!canSubmit || submitting) return;
        setSubmitting(true);
        setError(null);
        try {
            await verifyMFALogin(code, trustDevice);
            await completeSignIn(navigate, refreshUser);
        } catch (err) {
            if (err instanceof MFALoginError) {
                setError(err.code === 'MFA_CODE_INVALID' ? 'That code is incorrect. Try again.' : err.message);
                setSignInExpired(err.requiresNewSignIn);
            } else {
                setError('Verification failed. Please try again.');
            }
            setCode('');
            setSubmitting(false);
        }
    };

    const switchCodeType = () => {
        setUseBackupCode(current => !current);
        setCode('');
        setError(null);
    };

    return (
        <Container className='min-h-screen flex items-center justify-center py-12'>
            <Card className='w-full max-w-md'>
                <CardHeader>
                    <h1 className='text-2xl font-bold'>Two-factor authentication</h1>
                </CardHeader>
                <CardBody>
                    {signInExpired ?
                        <div className='space-y-4'>
                            <Alert variant='warning'>{error}</Alert>
                            <Button asChild variant='primary' className='w-full'>
                                <Link to='/login'>Sign in again</Link>
                            </Button>
                        </div>
                    :   <form onSubmit={handleSubmit} className='space-y-4' noValidate>
                            <p className='text-muted-foreground'>
                                {useBackupCode ?
                                    'Enter one of the 8-character backup codes you saved when you set up two-factor authentication. Each code works once.'
                                :   'Enter the 6-digit code from your authenticator app.'}
                            </p>
                            {error && <Alert variant='error'>{error}</Alert>}
                            <Input
                                label={useBackupCode ? 'Backup code' : 'Authentication code'}
                                name='mfa-code'
                                value={code}
                                onChange={event => setCode(event.target.value)}
                                inputMode={useBackupCode ? 'text' : 'numeric'}
                                autoComplete='one-time-code'
                                autoCapitalize={useBackupCode ? 'characters' : 'off'}
                                autoFocus
                                maxLength={useBackupCode ? 9 : 7}
                                placeholder={useBackupCode ? 'ABCD1234' : '123456'}
                                disabled={submitting}
                            />
                            <Checkbox
                                label='Trust this browser for 30 days'
                                checked={trustDevice}
                                onChange={event => setTrustDevice(event.target.checked)}
                                disabled={submitting}
                            />
                            <Button type='submit' variant='primary' className='w-full' loading={submitting} disabled={!canSubmit}>
                                Verify
                            </Button>
                            <button
                                type='button'
                                onClick={switchCodeType}
                                className='text-sm text-brand underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring'
                            >
                                {useBackupCode ? 'Use your authenticator app instead' : 'Use a backup code instead'}
                            </button>
                        </form>
                    }
                </CardBody>
            </Card>
        </Container>
    );
}
