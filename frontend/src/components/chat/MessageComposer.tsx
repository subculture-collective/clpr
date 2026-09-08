import React, { useState, useRef } from 'react';
import { Send } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { EmojiPicker } from './EmojiPicker';
import { searchUsersAutocomplete, type UserSuggestion } from '@/lib/user-api';

interface MessageComposerProps {
  onSend: (content: string) => void;
  onTyping?: () => void;
  placeholder?: string;
  disabled?: boolean;
}

// Note: Username pattern uses \w (alphanumeric and underscore only)
// Platforms supporting hyphens/periods in usernames should extend this pattern
const MENTION_PATTERN = /@(\w*)$/;

/**
 * MessageComposer component with emoji picker and mention autocomplete
 */
export function MessageComposer({
  onSend,
  onTyping,
  placeholder = 'Type a message...',
  disabled = false,
}: MessageComposerProps) {
  const [inputValue, setInputValue] = useState('');
  const [showMentions, setShowMentions] = useState(false);
  const [mentionSuggestions, setMentionSuggestions] = useState<UserSuggestion[]>([]);
  const [selectedMentionIndex, setSelectedMentionIndex] = useState(0);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const isTypingSentRef = useRef(false);
  const autocompleteTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  // Handle input change and detect mentions
  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setInputValue(value);

    // Trigger typing indicator with proper throttling
    if (onTyping && !isTypingSentRef.current) {
      onTyping();
      isTypingSentRef.current = true;
      
      // Clear existing timeout
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      
      // Reset throttle after 3 seconds
      typingTimeoutRef.current = setTimeout(() => {
        isTypingSentRef.current = false;
      }, 3000);
    }

    // Check for mention (@username)
    const cursorPosition = e.target.selectionStart;
    const textBeforeCursor = value.slice(0, cursorPosition);
    const mentionMatch = textBeforeCursor.match(MENTION_PATTERN);

    if (mentionMatch) {
      setShowMentions(true);
      setSelectedMentionIndex(0);
      
      // Clear existing autocomplete timeout
      if (autocompleteTimeoutRef.current) {
        clearTimeout(autocompleteTimeoutRef.current);
      }
      
      // Debounce autocomplete API calls by 300ms
      const query = mentionMatch[1];
      autocompleteTimeoutRef.current = setTimeout(() => {
        searchUsersAutocomplete(query, 10)
          .then(setMentionSuggestions)
          .catch((error) => {
            console.error('Failed to fetch user suggestions:', error);
            setMentionSuggestions([]);
          });
      }, 300);
    } else {
      setShowMentions(false);
      // Clear timeout if user stops typing @mention
      if (autocompleteTimeoutRef.current) {
        clearTimeout(autocompleteTimeoutRef.current);
      }
    }

    // Auto-resize textarea
    if (inputRef.current) {
      inputRef.current.style.height = 'auto';
      inputRef.current.style.height = `${Math.min(inputRef.current.scrollHeight, 120)}px`;
    }
  };

  // Handle sending message
  const handleSend = () => {
    if (!inputValue.trim() || disabled) return;
    
    onSend(inputValue);
    setInputValue('');
    setShowMentions(false);
    
    // Reset textarea height
    if (inputRef.current) {
      inputRef.current.style.height = 'auto';
    }
  };

  // Handle key down events
  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // Handle mention selection with arrow keys
    if (showMentions && mentionSuggestions.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedMentionIndex((prev) =>
          prev < mentionSuggestions.length - 1 ? prev + 1 : 0
        );
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedMentionIndex((prev) =>
          prev > 0 ? prev - 1 : mentionSuggestions.length - 1
        );
        return;
      }
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        insertMention(mentionSuggestions[selectedMentionIndex]);
        return;
      }
      if (e.key === 'Escape') {
        setShowMentions(false);
        return;
      }
    }

    // Send message on Enter (without shift)
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  // Insert mention into input
  const insertMention = (suggestion: UserSuggestion) => {
    const cursorPosition = inputRef.current?.selectionStart || 0;
    const textBeforeCursor = inputValue.slice(0, cursorPosition);
    const textAfterCursor = inputValue.slice(cursorPosition);
    const mentionStart = textBeforeCursor.lastIndexOf('@');
    
    const newValue =
      textBeforeCursor.slice(0, mentionStart) +
      `@${suggestion.username} ` +
      textAfterCursor;
    
    setInputValue(newValue);
    setShowMentions(false);
    inputRef.current?.focus();
  };

  // Handle emoji selection
  const handleEmojiSelect = (emoji: string) => {
    const cursorPosition = inputRef.current?.selectionStart || inputValue.length;
    const newValue =
      inputValue.slice(0, cursorPosition) +
      emoji +
      inputValue.slice(cursorPosition);
    setInputValue(newValue);
    inputRef.current?.focus();
  };

  return (
    <div className="border-t border-border bg-background p-4">
      {/* Mention suggestions */}
      {showMentions && mentionSuggestions.length > 0 && (
        <div className="mb-2 bg-background border border-border rounded-lg shadow-lg overflow-hidden">
          {mentionSuggestions.map((suggestion, index) => (
            <button
              key={suggestion.username}
              type="button"
              onClick={() => insertMention(suggestion)}
              className={`w-full text-left px-3 py-2 hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors ${
                index === selectedMentionIndex
                  ? 'bg-neutral-100 dark:bg-neutral-800'
                  : ''
              }`}
            >
              <div className="font-medium text-sm">@{suggestion.username}</div>
              {suggestion.display_name && (
                <div className="text-xs text-muted-foreground">
                  {suggestion.display_name}
                </div>
              )}
            </button>
          ))}
        </div>
      )}

      {/* Message input */}
      <div className="flex gap-2 items-end">
        <div className="flex-1 relative">
          <textarea
            ref={inputRef}
            value={inputValue}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            rows={1}
            className="w-full px-3 py-2.5 border border-border rounded-lg bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed resize-none min-h-[44px] max-h-[120px]"
            aria-label="Message input"
          />
        </div>
        
        <EmojiPicker onSelect={handleEmojiSelect} />
        
        <Button
          onClick={handleSend}
          disabled={!inputValue.trim() || disabled}
          size="sm"
          aria-label="Send message"
          className="px-4"
        >
          <Send className="w-4 h-4" />
          <span className="ml-2 hidden sm:inline">Send</span>
        </Button>
      </div>

      <div className="mt-2 text-xs text-muted-foreground">
        <kbd className="px-1.5 py-0.5 bg-neutral-100 dark:bg-neutral-800 rounded border border-border">
          Enter
        </kbd>{' '}
        to send,{' '}
        <kbd className="px-1.5 py-0.5 bg-neutral-100 dark:bg-neutral-800 rounded border border-border">
          Shift + Enter
        </kbd>{' '}
        for new line
      </div>
    </div>
  );
}
