import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export function Footer() {
    const { t } = useTranslation();
    const currentYear = new Date().getFullYear();

    return (
        <footer className='bg-background border-t border-border mt-auto'>
            <div className='container mx-auto px-4 py-8'>
                <div className='grid grid-cols-1 md:grid-cols-4 gap-8'>
                    {/* About Section */}
                    <div>
                        <h3 className='font-semibold mb-4'>
                            {t('footer.about')}
                        </h3>
                        <ul className='space-y-2'>
                            <li>
                                <Link
                                    to='/about'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.aboutClpr')}
                                </Link>
                            </li>
                            <li>
                                <a
                                    href='https://git.subcult.tv/subculture-collective/clpr'
                                    target='_blank'
                                    rel='noopener noreferrer'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.githubRepo')}
                                </a>
                            </li>
                        </ul>
                    </div>

                    {/* Legal Section */}
                    <div>
                        <h3 className='font-semibold mb-4'>
                            {t('footer.legal')}
                        </h3>
                        <ul className='space-y-2'>
                            <li>
                                <Link
                                    to='/privacy'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.privacyPolicy')}
                                </Link>
                            </li>
                            <li>
                                <Link
                                    to='/terms'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.termsOfService')}
                                </Link>
                            </li>
                            <li>
                                <Link
                                    to='/legal/dmca'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.dmcaPolicy')}
                                </Link>
                            </li>
                        </ul>
                    </div>

                    {/* Community Section */}
                    <div>
                        <h3 className='font-semibold mb-4'>
                            {t('footer.community')}
                        </h3>
                        <ul className='space-y-2'>
                            <li>
                                <Link
                                    to='/community-rules'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.communityRules')}
                                </Link>
                            </li>
                            <li>
                                <a
                                    href='https://discord.gg/TFwB4aJRef'
                                    target='_blank'
                                    rel='noopener noreferrer'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.discord')}
                                </a>
                            </li>
                            <li>
                                <a
                                    href='https://x.com/clpr_tv'
                                    target='_blank'
                                    rel='noopener noreferrer'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.twitter')}
                                </a>
                            </li>
                        </ul>
                    </div>

                    {/* Resources Section */}
                    <div>
                        <h3 className='font-semibold mb-4'>
                            {t('footer.resources')}
                        </h3>
                        <ul className='space-y-2'>
                            {/* <li>
                <Link
                  to="/docs"
                  className="text-muted-foreground hover:text-foreground transition-colors"
                >
                  Documentation
                </Link>
              </li> */}
                            <li>
                                <Link
                                    to='/contact'
                                    className='text-muted-foreground hover:text-foreground transition-colors'
                                >
                                    {t('footer.contactUs')}
                                </Link>
                            </li>
                        </ul>
                    </div>
                </div>

                {/* Copyright */}
                <div className='mt-8 pt-8 border-t border-border text-center text-muted-foreground'>
                    <p>
                        © {currentYear}{' '}
                        <a
                            href='https://subcult.tv'
                            target='_blank'
                            rel='noopener noreferrer'
                        >
                            subcult.tv
                        </a>{' '}
                        💜
                    </p>
                </div>
            </div>
        </footer>
    );
}
