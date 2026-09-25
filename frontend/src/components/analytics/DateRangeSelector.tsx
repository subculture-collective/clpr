import React from 'react';

interface DateRangeSelectorProps {
  value: number;
  onChange: (days: number) => void;
  options?: { label: string; value: number }[];
}

const DEFAULT_OPTIONS = [
  { label: '7 Days', value: 7 },
  { label: '30 Days', value: 30 },
  { label: '90 Days', value: 90 },
  { label: '1 Year', value: 365 },
];

const DateRangeSelector: React.FC<DateRangeSelectorProps> = ({
  value,
  onChange,
  options = DEFAULT_OPTIONS,
}) => {
  return (
    <div
      className="inline-flex max-w-full flex-wrap border border-line-strong"
      role="group"
      aria-label="Date range selection"
    >
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          onClick={() => onChange(option.value)}
          aria-pressed={value === option.value}
          className={`
            min-h-11 px-3 py-2 text-sm font-medium whitespace-nowrap
            ${
              value === option.value
                ? 'bg-primary-400 text-background'
                : 'bg-surface text-text-secondary hover:bg-surface-hover hover:text-foreground'
            }
            transition-colors duration-150 motion-reduce:transition-none
            focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2
          `}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
};

export default DateRangeSelector;
