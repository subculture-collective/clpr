import type { LucideIcon } from 'lucide-react';
import {
    Gamepad2,
    Swords,
    Trophy,
    Users,
    Laugh,
    Music,
    Film,
    Palette,
    BookOpen,
    Sparkles,
    Zap,
    Heart,
    Globe,
    Flame,
    Star,
    MessageSquare,
    Tv,
    Radio,
    Clapperboard,
    Rocket,
    Target,
    Crown,
    Dice5,
    Cpu,
    Mountain,
    Dumbbell,
    GraduationCap,
    Megaphone,
    Shapes,
    Newspaper,
    Landmark,
    Camera,
    Skull,
    Brain,
    Car,
    Ghost,
    Medal,
    Brush,
    Search,
    Tag,
} from 'lucide-react';

/**
 * Maps Lucide icon name strings (as stored in the database) to components.
 * Migration 000113 stores these as lowercase hyphenated names.
 */
const NAME_TO_ICON: Record<string, LucideIcon> = {
    newspaper: Newspaper,
    landmark: Landmark,
    flame: Flame,
    camera: Camera,
    trophy: Trophy,
    star: Star,
    skull: Skull,
    gamepad: Gamepad2,
    gamepad2: Gamepad2,
    swords: Swords,
    target: Target,
    crown: Crown,
    dice: Dice5,
    users: Users,
    'message-square': MessageSquare,
    heart: Heart,
    fire: Flame,
    music: Music,
    film: Film,
    palette: Palette,
    tv: Tv,
    radio: Radio,
    rocket: Rocket,
    cpu: Cpu,
    globe: Globe,
    mountain: Mountain,
    brain: Brain,
    car: Car,
    ghost: Ghost,
    medal: Medal,
    brush: Brush,
    search: Search,
    tag: Tag,
    sparkles: Sparkles,
    zap: Zap,
    megaphone: Megaphone,
    'book-open': BookOpen,
    'graduation-cap': GraduationCap,
};

/**
 * Maps known category emoji strings to Lucide icons.
 */
const EMOJI_TO_ICON: Record<string, LucideIcon> = {
    // Gaming
    '🎮': Gamepad2,
    '🕹️': Gamepad2,
    '⚔️': Swords,
    '🏆': Trophy,
    '🎯': Target,
    '👑': Crown,
    '🎲': Dice5,
    '🏅': Medal,

    // Social / Community
    '👥': Users,
    '😂': Laugh,
    '💬': MessageSquare,
    '❤️': Heart,
    '🔥': Flame,
    '⭐': Star,
    '📣': Megaphone,
    '🗳️': Landmark,

    // Media / Creative
    '🎵': Music,
    '🎶': Music,
    '🎬': Film,
    '📹': Film,
    '🎥': Clapperboard,
    '🎨': Palette,
    '🖌️': Brush,
    '📺': Tv,
    '📻': Radio,
    '📰': Newspaper,
    '📷': Camera,

    // Other categories
    '📚': BookOpen,
    '✨': Sparkles,
    '⚡': Zap,
    '🌍': Globe,
    '🌎': Globe,
    '🌟': Sparkles,
    '🚀': Rocket,
    '💻': Cpu,
    '🏔️': Mountain,
    '💪': Dumbbell,
    '🎓': GraduationCap,
    '🧩': Shapes,
    '🧠': Brain,
    '🚗': Car,
    '👻': Ghost,
};

interface CategoryIconProps {
    /** The icon value from category data — can be an emoji string or a Lucide icon name */
    icon?: string;
    /** Size variant */
    size?: 'sm' | 'md' | 'lg';
    /** Additional class names */
    className?: string;
}

const sizeConfig = {
    sm: { container: 'w-6 h-6', icon: 14, emoji: 'text-sm' },
    md: { container: 'w-8 h-8', icon: 18, emoji: 'text-lg' },
    lg: { container: 'w-12 h-12', icon: 28, emoji: 'text-3xl' },
};

export function CategoryIcon({ icon, size = 'md', className = '' }: CategoryIconProps) {
    if (!icon) return null;

    const config = sizeConfig[size];

    // Try Lucide icon name lookup first (e.g. "newspaper", "flame", "camera")
    const NamedIcon = NAME_TO_ICON[icon.toLowerCase()];
    if (NamedIcon) {
        return (
            <span
                className={`inline-flex items-center justify-center rounded-lg bg-brand/10 text-brand ${config.container} ${className}`}
                aria-hidden='true'
            >
                <NamedIcon size={config.icon} strokeWidth={1.75} />
            </span>
        );
    }

    // Try emoji lookup
    const EmojiIcon = EMOJI_TO_ICON[icon];
    if (EmojiIcon) {
        return (
            <span
                className={`inline-flex items-center justify-center rounded-lg bg-brand/10 text-brand ${config.container} ${className}`}
                aria-hidden='true'
            >
                <EmojiIcon size={config.icon} strokeWidth={1.75} />
            </span>
        );
    }

    // Fallback: render the raw value in a styled container
    return (
        <span
            className={`inline-flex items-center justify-center rounded-lg bg-brand/10 text-brand ${config.container} ${config.emoji} leading-none ${className}`}
            aria-hidden='true'
        >
            {icon}
        </span>
    );
}
