import React from "react";
import { Link } from "react-router-dom";
import type { Tag } from "../../types/tag";

interface TagChipProps {
  tag: Tag;
  size?: "small" | "medium";
  removable?: boolean;
  onRemove?: (slug: string) => void;
  onClick?: (slug: string) => void;
}

export const TagChip: React.FC<TagChipProps> = ({
  tag,
  size = "medium",
  removable = false,
  onRemove,
  onClick,
}) => {
  const sizeClasses = {
    small: "text-xs px-2 py-0.5",
    medium: "text-sm px-3 py-1",
  };

  const baseClasses = `inline-flex items-center gap-1 rounded-full font-medium transition-all hover:underline underline-offset-2 cursor-pointer ${sizeClasses[size]}`;

  // Compare true linearized sRGB luminance, rather than brightness, so
  // mid-tone provider colors also receive readable text.
  const hex = tag.color?.match(/^#([0-9a-f]{6})$/i)?.[1] || '6D28D9';
  const channels = [0, 2, 4].map(offset => {
    const channel = parseInt(hex.slice(offset, offset + 2), 16) / 255;
    return channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
  });
  const luminance = channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
  const style: React.CSSProperties = {
    backgroundColor: '#' + hex,
    color: luminance > 0.179 ? '#000000' : '#ffffff',
  };

  const handleClick = (e: React.MouseEvent) => {
    if (onClick) {
      e.preventDefault();
      onClick(tag.slug);
    }
  };

  const handleRemove = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (onRemove) {
      onRemove(tag.slug);
    }
  };

  const label = onClick ? (
    <button type='button' onClick={handleClick} className={baseClasses} style={style}>{tag.name}</button>
  ) : (
    <Link to={`/tags/${encodeURIComponent(tag.slug)}`} className={baseClasses} style={style}
      title={tag.description || `View clips tagged with ${tag.name}`}>{tag.name}</Link>
  );

  if (!removable) return label;
  return (
    <span className='inline-flex items-center rounded-full' style={style}>
      {label}
      <button type='button' onClick={handleRemove}
        className='mr-1 flex min-h-8 min-w-8 items-center justify-center rounded-full hover:bg-black/10'
        aria-label={`Remove ${tag.name} tag`}>
        <svg aria-hidden='true' className='h-3 w-3' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
          <path strokeLinecap='round' strokeLinejoin='round' strokeWidth={2} d='M6 18L18 6M6 6l12 12' />
        </svg>
      </button>
    </span>
  );
};
