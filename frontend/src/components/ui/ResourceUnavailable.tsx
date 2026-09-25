import { Link } from 'react-router-dom';
import { Button } from './Button';

interface ResourceLink {
    label: string;
    href: string;
}

interface ResourceUnavailableProps {
    /** `not-found` for a 404; `error` for any other failure. */
    kind: 'not-found' | 'error';
    title: string;
    description: string;
    /** Shown as "Try again" when provided. Only meaningful for `error`. */
    onRetry?: () => void;
    /** Routes that give the visitor somewhere useful to go next. */
    links?: ResourceLink[];
    headingLevel?: 'h1' | 'h2';
    /** Overrides the small label above the title. */
    kicker?: string;
}

/**
 * Page-level state for a detail route whose resource is missing or failed to
 * load. Never pass raw transport messages (e.g. axios `error.message`) here.
 */
export function ResourceUnavailable({
    kind,
    title,
    description,
    onRetry,
    links = [],
    headingLevel: Heading = 'h1',
    kicker,
}: ResourceUnavailableProps) {
    return (
        <section
            role={kind === 'error' ? 'alert' : undefined}
            aria-labelledby='resource-unavailable-title'
            className='mx-auto my-8 max-w-xl border border-line-strong bg-surface p-6 text-center tally-bar sm:p-8'
            data-testid='resource-unavailable'
        >
            <p className='kicker mb-3'>
                {kicker ?? (kind === 'not-found' ? 'Not found' : 'Something went wrong')}
            </p>
            <Heading id='resource-unavailable-title' className='mb-3 text-3xl text-foreground'>
                {title}
            </Heading>
            <p className='mb-6 text-text-secondary'>{description}</p>
            {(onRetry || links.length > 0) && (
                <div className='flex flex-wrap justify-center gap-3'>
                    {onRetry && (
                        <Button onClick={onRetry}>Try again</Button>
                    )}
                    {links.map((link, index) => (
                        <Button
                            key={link.href}
                            asChild
                            variant={!onRetry && index === 0 ? 'primary' : 'secondary'}
                        >
                            <Link to={link.href}>{link.label}</Link>
                        </Button>
                    ))}
                </div>
            )}
        </section>
    );
}
