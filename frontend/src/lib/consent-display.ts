/** Returns the effective optional-consent state without mutating the saved choice. */
export function effectiveConsentValue(stored: boolean, doNotTrack: boolean) {
    return doNotTrack ? false : stored;
}
