import type { Config } from "tailwindcss";

// all in fixtures is set to tailwind v3 as interims solutions

const config: Config = {
    darkMode: ["class"],
    content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "*.{js,ts,jsx,tsx,mdx}"
  ],
  theme: {
  	extend: {
  		colors: {
  			background: 'hsl(var(--background))',
  			foreground: 'hsl(var(--foreground))',
  			card: {
  				DEFAULT: 'hsl(var(--card))',
  				foreground: 'hsl(var(--card-foreground))'
  			},
  			popover: {
  				DEFAULT: 'hsl(var(--popover))',
  				foreground: 'hsl(var(--popover-foreground))'
  			},
  			primary: {
  				DEFAULT: 'hsl(var(--primary))',
  				foreground: 'hsl(var(--primary-foreground))'
  			},
  			secondary: {
  				DEFAULT: 'hsl(var(--secondary))',
  				foreground: 'hsl(var(--secondary-foreground))'
  			},
  			muted: {
  				DEFAULT: 'hsl(var(--muted))',
  				foreground: 'hsl(var(--muted-foreground))'
  			},
  			accent: {
  				DEFAULT: 'hsl(var(--accent))',
  				foreground: 'hsl(var(--accent-foreground))',
  				cyan: 'hsl(var(--accent-cyan))',
  				'cyan-light': 'hsl(var(--accent-cyan-light))',
  				'cyan-dark': 'hsl(var(--accent-cyan-dark))',
  				teal: 'hsl(var(--accent-teal))',
  				'teal-light': 'hsl(var(--accent-teal-light))',
  				'teal-dark': 'hsl(var(--accent-teal-dark))',
  				amber: 'hsl(var(--accent-amber))',
  				'amber-light': 'hsl(var(--accent-amber-light))',
  				'amber-dark': 'hsl(var(--accent-amber-dark))',
  			},
  			destructive: {
  				DEFAULT: 'hsl(var(--destructive))',
  				foreground: 'hsl(var(--destructive-foreground))'
  			},
  			border: 'hsl(var(--border))',
  			input: 'hsl(var(--input))',
  			ring: 'hsl(var(--ring))',
  			chart: {
  				'1': 'hsl(var(--chart-1))',
  				'2': 'hsl(var(--chart-2))',
  				'3': 'hsl(var(--chart-3))',
  				'4': 'hsl(var(--chart-4))',
  				'5': 'hsl(var(--chart-5))'
  			},
  			sidebar: {
  				DEFAULT: 'hsl(var(--sidebar-background))',
  				foreground: 'hsl(var(--sidebar-foreground))',
  				primary: 'hsl(var(--sidebar-primary))',
  				'primary-foreground': 'hsl(var(--sidebar-primary-foreground))',
  				accent: 'hsl(var(--sidebar-accent))',
  				'accent-foreground': 'hsl(var(--sidebar-accent-foreground))',
  				border: 'hsl(var(--sidebar-border))',
  				ring: 'hsl(var(--sidebar-ring))'
  			},
  			// Academic Horizon Color Palette
  			'primary-blue': {
  				50: 'hsl(var(--primary-blue-50))',
  				100: 'hsl(var(--primary-blue-100))',
  				200: 'hsl(var(--primary-blue-200))',
  				300: 'hsl(var(--primary-blue-300))',
  				400: 'hsl(var(--primary-blue-400))',
  				500: 'hsl(var(--primary-blue-500))',
  				600: 'hsl(var(--primary-blue-600))',
  				700: 'hsl(var(--primary-blue-700))',
  				800: 'hsl(var(--primary-blue-800))',
  				900: 'hsl(var(--primary-blue-900))',
  			},
  			'secondary-purple': {
  				50: 'hsl(var(--secondary-purple-50))',
  				100: 'hsl(var(--secondary-purple-100))',
  				200: 'hsl(var(--secondary-purple-200))',
  				300: 'hsl(var(--secondary-purple-300))',
  				400: 'hsl(var(--secondary-purple-400))',
  				500: 'hsl(var(--secondary-purple-500))',
  				600: 'hsl(var(--secondary-purple-600))',
  				700: 'hsl(var(--secondary-purple-700))',
  				800: 'hsl(var(--secondary-purple-800))',
  				900: 'hsl(var(--secondary-purple-900))',
  			},
  			bg: {
  				DEFAULT: 'hsl(var(--background))',
  				base: 'hsl(var(--bg-base))',
  				elevation: 'hsl(var(--bg-elevation))',
  				glass: 'hsl(var(--bg-glass))',
				accent: 'hsl(var(--bg-accent))',
  				'glass-hover': 'hsl(var(--bg-glass-hover))',
  				'glass-elevated': 'hsl(var(--bg-glass-elevated))',
  			},
  			glass: {
  				bg: 'hsl(var(--glass-bg))',
  				border: 'hsl(var(--glass-border))',
  				highlight: 'hsl(var(--glass-highlight))',
  				shadow: 'hsl(var(--glass-shadow))',
  			},
  			text: {
  				title: 'hsl(var(--text-title))',
  				caption: 'hsl(var(--text-caption))',
  				body: 'hsl(var(--text-body))',
  				muted: 'hsl(var(--text-muted))',
  				disabled: 'hsl(var(--text-disabled))',
  			},
  			status: {
  				success: 'hsl(var(--status-success))',
  				error: 'hsl(var(--status-error))',
  				warning: 'hsl(var(--status-warning))',
  				info: 'hsl(var(--status-info))',
  			},
  			link: {
  				DEFAULT: 'hsl(var(--link-default))',
  				hover: 'hsl(var(--link-hover))',
  				active: 'hsl(var(--link-active))',
  			},
  			'border-focus': 'hsl(var(--border-focus))',
  			'border-divider': 'hsl(var(--border-divider))',
  		},
  		borderRadius: {
  			lg: 'var(--radius)',
  			md: 'calc(var(--radius) - 2px)',
  			sm: 'calc(var(--radius) - 4px)'
  		},
  		keyframes: {
  			'accordion-down': {
  				from: {
  					height: '0'
  				},
  				to: {
  					height: 'var(--radix-accordion-content-height)'
  				}
  			},
  			'accordion-up': {
  				from: {
  					height: 'var(--radix-accordion-content-height)'
  				},
  				to: {
  					height: '0'
  				}
  			}
  		},
  		animation: {
  			'accordion-down': 'accordion-down 0.2s ease-out',
  			'accordion-up': 'accordion-up 0.2s ease-out'
  		}
  	}
  },
  plugins: [require("tailwindcss-animate")],
};
export default config;
