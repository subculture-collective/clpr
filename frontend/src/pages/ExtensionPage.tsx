import { Link } from 'react-router-dom';
import { Search, MousePointer, Tag, PenLine, Zap, Lock } from 'lucide-react';
import { Container, Card, CardBody, SEO, Button } from '../components';

const CHROME_STORE_URL = 'https://chrome.google.com/webstore/detail/clipper';
const FIREFOX_STORE_URL = 'https://addons.mozilla.org/firefox/addon/clipper';
const GITHUB_EXTENSION_URL =
    'https://github.com/subculture-collective/clipper/tree/main/extension';

interface FeatureProps {
    icon: React.ReactNode;
    title: string;
    description: string;
}

function Feature({ icon, title, description }: FeatureProps) {
    return (
        <div className="flex gap-4">
            <span className="flex-shrink-0" aria-hidden="true">
                {icon}
            </span>
            <div>
                <h3 className="font-semibold mb-1">{title}</h3>
                <p className="text-sm text-muted-foreground">{description}</p>
            </div>
        </div>
    );
}

export function ExtensionPage() {
    return (
        <>
            <SEO
                title="Browser Extension"
                description="Share Twitch clips to Clipper with one click. Get the Clipper browser extension for Chrome and Firefox."
                canonicalUrl="/extension"
            />
            <Container className="py-8 max-w-4xl">
                {/* Hero */}
                <div className="mb-10 text-center">
                    <h1 className="text-4xl font-bold mb-4">
                        Clipper Browser Extension
                    </h1>
                    <p className="text-lg text-muted-foreground mb-8 max-w-2xl mx-auto">
                        Share Twitch clips to Clipper with one click. The extension
                        detects clips automatically, pre-fills metadata, and lets you
                        add tags and a description before submitting.
                    </p>
                    <div className="flex flex-wrap gap-3 justify-center">
                        <a
                            href={CHROME_STORE_URL}
                            target="_blank"
                            rel="noopener noreferrer"
                            aria-label="Get Clipper for Chrome"
                        >
                            <Button variant="primary" size="lg">
                                Add to Chrome
                            </Button>
                        </a>
                        <a
                            href={FIREFOX_STORE_URL}
                            target="_blank"
                            rel="noopener noreferrer"
                            aria-label="Get Clipper for Firefox"
                        >
                            <Button variant="secondary" size="lg">
                                Add to Firefox
                            </Button>
                        </a>
                    </div>
                </div>

                {/* Features */}
                <Card className="mb-8">
                    <CardBody>
                        <h2 className="text-2xl font-semibold mb-6">Features</h2>
                        <div className="grid gap-6 sm:grid-cols-2">
                            <Feature
                                icon={<Search size={16} strokeWidth={1.75} />}
                                title="Auto-detect clips"
                                description="Automatically detects Twitch clip pages (twitch.tv and clips.twitch.tv) and enables the share button."
                            />
                            <Feature
                                icon={<MousePointer size={16} strokeWidth={1.75} />}
                                title="Context menu"
                                description='Right-click any Twitch clip page to see "Share to Clipper" in the context menu.'
                            />
                            <Feature
                                icon={<PenLine size={16} strokeWidth={1.75} />}
                                title="Editable metadata"
                                description="Pre-fills the clip title from Twitch. You can edit the title, add a description, and pick tags before sharing."
                            />
                            <Feature
                                icon={<Tag size={16} strokeWidth={1.75} />}
                                title="Tag selection"
                                description="Browse and search all Clipper tags directly in the popup and apply multiple tags to your submission."
                            />
                            <Feature
                                icon={<Zap size={16} strokeWidth={1.75} />}
                                title="One-click submit"
                                description="Click Share Clip to submit instantly. A desktop notification confirms when your clip is pending review."
                            />
                            <Feature
                                icon={<Lock size={16} strokeWidth={1.75} />}
                                title="Secure auth"
                                description="Authenticates using your existing Clipper account. No separate credentials required."
                            />
                        </div>
                    </CardBody>
                </Card>

                {/* How it works */}
                <Card className="mb-8">
                    <CardBody>
                        <h2 className="text-2xl font-semibold mb-6">How it works</h2>
                        <ol className="space-y-4 list-decimal list-inside text-sm text-muted-foreground">
                            <li>
                                <strong className="text-foreground">Install</strong> the
                                extension from the Chrome Web Store or Firefox Add-ons.
                            </li>
                            <li>
                                <strong className="text-foreground">Log in</strong> by
                                clicking the extension icon and selecting{' '}
                                <em>Login with Twitch</em>. This opens your Clipper account
                                in a new tab.
                            </li>
                            <li>
                                <strong className="text-foreground">Browse Twitch</strong>.
                                When you land on a clip page the extension badge lights up
                                automatically.
                            </li>
                            <li>
                                <strong className="text-foreground">Click the icon</strong>{' '}
                                (or right-click → Share to Clipper) to open the popup.
                            </li>
                            <li>
                                <strong className="text-foreground">
                                    Review and submit
                                </strong>{' '}
                                – edit the title, add tags, and click <em>Share Clip</em>.
                            </li>
                        </ol>
                    </CardBody>
                </Card>

                {/* Supported browsers */}
                <Card className="mb-8">
                    <CardBody>
                        <h2 className="text-2xl font-semibold mb-4">
                            Supported browsers
                        </h2>
                        <div className="overflow-x-auto">
                            <table className="w-full text-sm">
                                <thead>
                                    <tr className="border-b border-border">
                                        <th className="text-left py-2 pr-8">Browser</th>
                                        <th className="text-left py-2">Minimum version</th>
                                    </tr>
                                </thead>
                                <tbody className="text-muted-foreground">
                                    <tr className="border-b border-border">
                                        <td className="py-2 pr-8">Chrome / Chromium</td>
                                        <td className="py-2">99+</td>
                                    </tr>
                                    <tr className="border-b border-border">
                                        <td className="py-2 pr-8">Microsoft Edge</td>
                                        <td className="py-2">99+</td>
                                    </tr>
                                    <tr>
                                        <td className="py-2 pr-8">Firefox</td>
                                        <td className="py-2">109+</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </CardBody>
                </Card>

                {/* Open source */}
                <Card>
                    <CardBody>
                        <h2 className="text-2xl font-semibold mb-4">Open source</h2>
                        <p className="text-muted-foreground mb-4">
                            The Clipper extension is open source and available on GitHub.
                            Contributions, bug reports, and feature requests are welcome.
                        </p>
                        <div className="flex flex-wrap gap-3">
                            <a
                                href={GITHUB_EXTENSION_URL}
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                <Button variant="secondary">View on GitHub</Button>
                            </a>
                            <Link to="/about">
                                <Button variant="ghost">About Clipper</Button>
                            </Link>
                        </div>
                    </CardBody>
                </Card>
            </Container>
        </>
    );
}
