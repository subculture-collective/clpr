export function resolveSiteUrl(
    value: string,
    configuredBase: string | undefined,
    runtimeOrigin: string,
): string {
    const siteOrigin =
        configuredBase && /^https?:\/\//i.test(configuredBase)
            ? configuredBase
            : runtimeOrigin;
    return new URL(value, siteOrigin).toString();
}
