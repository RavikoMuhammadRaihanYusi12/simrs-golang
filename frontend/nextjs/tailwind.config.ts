import type { Config } from "tailwindcss";

const colors = {
    primary: {
        DEFAULT: "#2563eb",
        hover: "#1d4ed8",
        active: "#1e40af",
        soft: "#dbebfd",
    },
    secondary: {
        DEFAULT: "#64748b",
        hover: "#475569",
    },
    accent: {
        DEFAULT: "#0d9488",
        hover: "#0f766e",
        soft: "#ccfbf1",
    },
    error: {
        DEFAULT: "#dc2626",
        hover: "#b91c1c",
        soft: "#fee2e2",
    },
    success: {
        DEFAULT: "#16a34a",
        hover: "#15803d",
        soft: "#dcfce7",
    },
    warning: {
        DEFAULT: "#d97706",
        soft: "#fef3c7",
    },
    background: "#f8fafc",
    foreground: "#0f172a",
    surface: "#ffffff",
    border: "#e2e8f0",
};

const spacing = {
    touch: "44px",
    section: "2rem",
};

export default {
    content: [
        "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
        "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
        "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    darkMode: "class",
    theme: {
        extend: {
            colors,
            spacing,
            fontFamily: {
                sans: [
                    "var(--font-inter)",
                    "ui-sans-serif",
                    "system-ui",
                    "-apple-system",
                    "BlinkMacSystemFont",
                    "Segoe UI",
                    "Roboto",
                    "sans-serif",
                ],
                display: [
                    "var(--font-jakarta)",
                    "var(--font-inter)",
                    "ui-sans-serif",
                    "system-ui",
                    "sans-serif",
                ],
            },
            boxShadow: {
                card: "0 1px 3px 0 rgb(15 23 42 / 0.06), 0 1px 2px -1px rgb(15 23 42 / 0.06)",
                "card-hover": "0 4px 12px -2px rgb(15 23 42 / 0.10)",
            },
            keyframes: {
                "fade-up": {
                    "0%": { opacity: "0", transform: "translateY(8px)" },
                    "100%": { opacity: "1", transform: "translateY(0)" },
                },
            },
            animation: {
                "fade-up": "fade-up 0.35s ease-out both",
            },
            borderRadius: {
                xl: "0.875rem",
            },
            lineHeight: {
                body: "1.6",
            },
        },
    },
    plugins: [],
} satisfies Config;
