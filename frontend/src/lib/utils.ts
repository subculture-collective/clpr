import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { formatDistanceToNow, format, differenceInMinutes } from 'date-fns';

/**
 * Utility function to merge Tailwind CSS classes
 * Uses clsx for conditional classes and tailwind-merge to resolve conflicts
 */
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

/**
 * Format timestamp with improved UX:
 * - Shows relative time (e.g., "2 minutes ago") for times within the last hour
 * - Shows exact time (e.g., "3:42 PM") for times after 1 hour
 * - Full date shown in title attribute for tooltip
 */
export function formatTimestamp(date: Date | string): {
    display: string;
    title: string;
} {
    const dateObj = typeof date === 'string' ? new Date(date) : date;

    // Validate date
    if (isNaN(dateObj.getTime())) {
        return {
            display: 'Invalid date',
            title: 'Invalid date',
        };
    }

    const now = new Date();
    const minutesDiff = differenceInMinutes(now, dateObj);

    // If the date is in the future, show exact time
    if (minutesDiff < 0) {
        const timeDisplay = format(dateObj, 'h:mm a');
        return {
            display: timeDisplay,
            title: format(dateObj, 'PPpp'),
        };
    }

    // For times within the last hour (including "1 hour ago"), show relative time
    if (minutesDiff <= 60) {
        const relative = formatDistanceToNow(dateObj, { addSuffix: true });
        return {
            display: relative,
            title: format(dateObj, 'PPpp'), // Full date and time
        };
    }

    // For times after 1 hour, show exact time with date context
    const fullTitle = format(dateObj, 'PPpp');
    const daysDiff = Math.floor(minutesDiff / (60 * 24));
    const yearsDiff = now.getFullYear() - dateObj.getFullYear();

    let timeDisplay: string;
    if (daysDiff === 0) {
        // Today: just show time
        timeDisplay = format(dateObj, 'h:mm a'); // e.g., "3:42 PM"
    } else if (daysDiff === 1) {
        // Yesterday
        timeDisplay = `Yesterday ${format(dateObj, 'h:mm a')}`;
    } else if (yearsDiff === 0) {
        // This year: show date and time
        timeDisplay = format(dateObj, 'MMM d, h:mm a'); // e.g., "Mar 15, 3:42 PM"
    } else {
        // Previous years: show full date
        timeDisplay = format(dateObj, 'MMM d, yyyy'); // e.g., "Mar 15, 2024"
    }

    return {
        display: timeDisplay,
        title: fullTitle,
    };
}

/**
 * Format a timestamp as a relative date for compact UI surfaces.
 * Examples: "2 hours ago", "3 days ago"
 */
export function formatRelativeTimestamp(date: Date | string): {
    display: string;
    title: string;
} {
    const dateObj = typeof date === 'string' ? new Date(date) : date;

    if (isNaN(dateObj.getTime())) {
        return {
            display: 'Invalid date',
            title: 'Invalid date',
        };
    }

    return {
        display: formatDistanceToNow(dateObj, { addSuffix: true }),
        title: format(dateObj, 'PPpp'),
    };
}

/**
 * List of valid CSS color names
 * Based on standard CSS3/CSS4 color keywords
 */
const CSS_COLOR_NAMES = new Set([
    'aliceblue',
    'antiquewhite',
    'aqua',
    'aquamarine',
    'azure',
    'beige',
    'bisque',
    'black',
    'blanchedalmond',
    'blue',
    'blueviolet',
    'brown',
    'burlywood',
    'cadetblue',
    'chartreuse',
    'chocolate',
    'coral',
    'cornflowerblue',
    'cornsilk',
    'crimson',
    'cyan',
    'darkblue',
    'darkcyan',
    'darkgoldenrod',
    'darkgray',
    'darkgrey',
    'darkgreen',
    'darkkhaki',
    'darkmagenta',
    'darkolivegreen',
    'darkorange',
    'darkorchid',
    'darkred',
    'darksalmon',
    'darkseagreen',
    'darkslateblue',
    'darkslategray',
    'darkslategrey',
    'darkturquoise',
    'darkviolet',
    'deeppink',
    'deepskyblue',
    'dimgray',
    'dimgrey',
    'dodgerblue',
    'firebrick',
    'floralwhite',
    'forestgreen',
    'fuchsia',
    'gainsboro',
    'ghostwhite',
    'gold',
    'goldenrod',
    'gray',
    'grey',
    'green',
    'greenyellow',
    'honeydew',
    'hotpink',
    'indianred',
    'indigo',
    'ivory',
    'khaki',
    'lavender',
    'lavenderblush',
    'lawngreen',
    'lemonchiffon',
    'lightblue',
    'lightcoral',
    'lightcyan',
    'lightgoldenrodyellow',
    'lightgray',
    'lightgrey',
    'lightgreen',
    'lightpink',
    'lightsalmon',
    'lightseagreen',
    'lightskyblue',
    'lightslategray',
    'lightslategrey',
    'lightsteelblue',
    'lightyellow',
    'lime',
    'limegreen',
    'linen',
    'magenta',
    'maroon',
    'mediumaquamarine',
    'mediumblue',
    'mediumorchid',
    'mediumpurple',
    'mediumseagreen',
    'mediumslateblue',
    'mediumspringgreen',
    'mediumturquoise',
    'mediumvioletred',
    'midnightblue',
    'mintcream',
    'mistyrose',
    'moccasin',
    'navajowhite',
    'navy',
    'oldlace',
    'olive',
    'olivedrab',
    'orange',
    'orangered',
    'orchid',
    'palegoldenrod',
    'palegreen',
    'paleturquoise',
    'palevioletred',
    'papayawhip',
    'peachpuff',
    'peru',
    'pink',
    'plum',
    'powderblue',
    'purple',
    'rebeccapurple',
    'red',
    'rosybrown',
    'royalblue',
    'saddlebrown',
    'salmon',
    'sandybrown',
    'seagreen',
    'seashell',
    'sienna',
    'silver',
    'skyblue',
    'slateblue',
    'slategray',
    'slategrey',
    'snow',
    'springgreen',
    'steelblue',
    'tan',
    'teal',
    'thistle',
    'tomato',
    'turquoise',
    'violet',
    'wheat',
    'white',
    'whitesmoke',
    'yellow',
    'yellowgreen',
    'transparent',
]);

/**
 * Sanitizes a color value for safe use in inline CSS styles.
 *
 * This provides defense-in-depth against CSS injection attacks by validating
 * that color values conform to expected CSS color formats.
 *
 * Supported formats:
 * - Hex colors: #RGB, #RRGGBB, #RRGGBBAA
 * - RGB/RGBA: rgb(r, g, b), rgba(r, g, b, a)
 * - HSL/HSLA: hsl(h, s%, l%), hsla(h, s%, l%, a)
 * - Named CSS colors (e.g., 'red', 'blue', 'transparent')
 *
 * @param color - The color value to sanitize
 * @returns The sanitized color value, or null if invalid
 *
 * @example
 * sanitizeColor('#ff0000') // '#ff0000'
 * sanitizeColor('rgb(255, 0, 0)') // 'rgb(255, 0, 0)'
 * sanitizeColor('red') // 'red'
 * sanitizeColor('javascript:alert(1)') // null
 * sanitizeColor('expression(alert(1))') // null
 */
export function sanitizeColor(color: string | null | undefined): string | null {
    if (!color || typeof color !== 'string') {
        return null;
    }

    // Trim and convert to lowercase for validation
    const trimmedColor = color.trim().toLowerCase();

    // Reject empty strings
    if (trimmedColor === '') {
        return null;
    }

    // Check for dangerous patterns that could be used for CSS injection
    const dangerousPatterns = [
        /[<>'"]/, // HTML/JS injection characters
        /javascript:/i, // JavaScript protocol
        /expression\(/i, // IE expression()
        /import/i, // CSS @import
        /url\(/i, // CSS url() - could load external resources
        /var\(/i, // CSS variables could be used for injection
        /calc\(/i, // CSS calc could be used for injection
        /attr\(/i, // CSS attr could be used for injection
        /;/, // CSS statement separator
        /\\/, // Backslash escaping
    ];

    for (const pattern of dangerousPatterns) {
        if (pattern.test(trimmedColor)) {
            return null;
        }
    }

    // Validate hex color format (#RGB, #RRGGBB, #RRGGBBAA)
    const hexPattern = /^#([0-9a-f]{3}|[0-9a-f]{6}|[0-9a-f]{8})$/;
    if (hexPattern.test(trimmedColor)) {
        return trimmedColor;
    }

    // Validate rgb/rgba format
    const rgbPattern =
        /^rgba?\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*(?:,\s*(0|1|0?\.\d+)\s*)?\)$/;
    const rgbMatch = trimmedColor.match(rgbPattern);
    if (rgbMatch) {
        const [, r, g, b, a] = rgbMatch;
        // Validate RGB values are in range 0-255
        const rVal = parseInt(r, 10);
        const gVal = parseInt(g, 10);
        const bVal = parseInt(b, 10);

        if (rVal <= 255 && gVal <= 255 && bVal <= 255) {
            // Validate alpha if present (0-1)
            if (a !== undefined) {
                const aVal = parseFloat(a);
                if (aVal >= 0 && aVal <= 1) {
                    return trimmedColor;
                }
            } else {
                return trimmedColor;
            }
        }
        return null;
    }

    // Validate hsl/hsla format
    const hslPattern =
        /^hsla?\(\s*(\d{1,3})\s*,\s*(\d{1,3})%\s*,\s*(\d{1,3})%\s*(?:,\s*(0|1|0?\.\d+)\s*)?\)$/;
    const hslMatch = trimmedColor.match(hslPattern);
    if (hslMatch) {
        const [, h, s, l, a] = hslMatch;
        // Validate HSL values are in range
        const hVal = parseInt(h, 10);
        const sVal = parseInt(s, 10);
        const lVal = parseInt(l, 10);

        if (hVal <= 360 && sVal <= 100 && lVal <= 100) {
            // Validate alpha if present (0-1)
            if (a !== undefined) {
                const aVal = parseFloat(a);
                if (aVal >= 0 && aVal <= 1) {
                    return trimmedColor;
                }
            } else {
                return trimmedColor;
            }
        }
        return null;
    }

    // Validate named CSS colors
    if (CSS_COLOR_NAMES.has(trimmedColor)) {
        return trimmedColor;
    }

    // If none of the patterns match, reject the color
    return null;
}

/**
 * Format duration in seconds to human-readable format (MM:SS or HH:MM:SS)
 * @param seconds - Duration in seconds
 * @returns Formatted duration string
 */
export function formatDuration(seconds: number): string {
    if (!seconds || seconds < 0) {
        return '0:00';
    }

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);

    if (hours > 0) {
        return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    }

    return `${minutes}:${secs.toString().padStart(2, '0')}`;
}

/** Compact, locale-stable counts for dense cards and controls. */
export function formatCompactNumber(value: number): string {
    if (!Number.isFinite(value)) return '0';
    if (Math.abs(value) >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
    if (Math.abs(value) >= 1_000) return `${(value / 1_000).toFixed(1)}K`;
    return Math.trunc(value).toString();
}
