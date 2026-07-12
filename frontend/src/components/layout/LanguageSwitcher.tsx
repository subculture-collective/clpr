import { useTranslation } from 'react-i18next';
import { Button } from '../ui/Button';

const languages = [
  { code: 'en', name: 'English', flag: '🇬🇧' },
  { code: 'es', name: 'Español', flag: '🇪🇸' },
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
];

export function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng);
  };

  return (
    <div className="flex gap-1">
      {languages.map((lang) => (
        <Button
          key={lang.code}
          variant="ghost"
          size="sm"
          onClick={() => changeLanguage(lang.code)}
          className={i18n.language === lang.code ? 'bg-primary-100 dark:bg-primary-900' : ''}
          title={lang.name}
        >
          {lang.flag}
        </Button>
      ))}
    </div>
  );
}
