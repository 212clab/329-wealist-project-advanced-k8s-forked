import React from 'react';
import { Palette } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';

const ThemeSelector: React.FC = () => {
  const { theme, themeName, setTheme } = useTheme();

  const themeOptions = [
    { name: 'modern' as const, label: 'MODERN', color: 'bg-blue-600' },
  ];

  return (
    <div className="relative inline-block">
      <div 
        className={`relative flex gap-1 sm:gap-2 p-1 ${theme.colors.secondary} ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.effects.borderRadius}`} 
        style={theme.name === 'retro' ? { boxShadow: theme.effects.shadow } : { boxShadow: theme.effects.shadow }}
      >
        {themeOptions.map((option) => (
          <button
            key={option.name}
            onClick={() => setTheme(option.name)}
            className={`px-2 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} transition ${theme.effects.borderRadius} ${
              themeName === option.name 
                ? `${option.color} text-white` 
                : `${theme.colors.secondary} ${theme.colors.text} hover:bg-gray-100`
            }`}
            title={`${option.label} Theme`}
          >
            {themeName === option.name ? (
              <div className={`w-4 h-4 ${option.color} ${theme.effects.cardBorderWidth} ${theme.colors.border}`}></div>
            ) : (
              <Palette className="w-4 h-4" />
            )}
          </button>
        ))}
      </div>
    </div>
  );
};

export default ThemeSelector;