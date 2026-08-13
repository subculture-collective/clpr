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

  const baseClasses = `inline-flex items-center gap-1 rounded-full font-medium transition-all hover:opacity-80 cursor-pointer ${sizeClasses[size]}`;

  const backgroundColor = tag.color || '#6D28D9';
  const hex = backgroundColor.match(/^#([0-9a-f]{6})$/i)?.[1];
  const luminance = hex
    ? (0.2126 * parseInt(hex.slice(0, 2), 16) +
        0.7152 * parseInt(hex.slice(2, 4), 16) +
        0.0722 * parseInt(hex.slice(4, 6), 16)) / 255
    : 0;
  const style: React.CSSProperties = {
    backgroundColor,
    color: luminance > 0.56 ? '#111318' : '#ffffff',
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

  const content = (
    <>
      <span>{tag.name}</span>
      {removable && (
        <button
          onClick={handleRemove}
          className="ml-1 hover:bg-white/20 rounded-full p-0.5 cursor-pointer"
          aria-label={`Remove ${tag.name} tag`}
        >
          <svg
            className="w-3 h-3"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      )}
    </>
  );

  if (onClick) {
    return (
      <button onClick={handleClick} className={baseClasses} style={style}>
        {content}
      </button>
    );
  }

  return (
    <Link
      to={`/tags/${encodeURIComponent(tag.slug)}`}
      className={baseClasses}
      style={style}
      title={tag.description || `View clips tagged with ${tag.name}`}
    >
      {content}
    </Link>
  );
};
