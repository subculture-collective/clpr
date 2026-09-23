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
      className="inline-flex rounded-md shadow-sm"
      role="group"
      aria-label="Date range selection"
    >
      {options.map((option, index) => (
        <button
          key={option.value}
          type="button"
          onClick={() => onChange(option.value)}
          aria-pressed={value === option.value}
          className={`
            px-4 py-2 text-sm font-medium
            ${index === 0 ? 'rounded-l-lg' : ''}
            ${index === options.length - 1 ? 'rounded-r-lg' : ''}
            ${
              value === option.value
                ? 'bg-primary-400 text-background'
                : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
            }
            border border-gray-200 dark:border-gray-600
            transition-colors duration-150
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
