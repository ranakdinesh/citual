"use client";

import Link from "next/link";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";

export function Header() {
    const { theme, setTheme } = useTheme();
    const [mounted, setMounted] = useState(false);
    const [isMenuOpen, setIsMenuOpen] = useState(false);

    useEffect(() => {
        setMounted(true);
    }, []);

    const toggleTheme = () => {
        setTheme(theme === "dark" ? "light" : "dark");
    };

    return (
        <header className="sticky top-0 z-50 w-full border-b border-white/10 dark:border-border-dark bg-bg-light/80 dark:bg-bg-dark/80 backdrop-blur-md transition-colors duration-300">
            <div className="container mx-auto flex h-16 items-center justify-between px-6 md:px-10">
                {/* Logo */}
                <div className="flex items-center gap-2">
                    <Link href="/" className="flex items-center gap-2 group">
                        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-white group-hover:bg-primary/90 transition-colors">
                            <span className="material-symbols-outlined text-[20px]">dashboard_customize</span>
                        </div>
                        <span className="text-lg font-bold tracking-tight text-slate-900 dark:text-white">
                            Citual
                        </span>
                    </Link>
                </div>

                {/* Desktop Navigation */}
                <nav className="hidden md:flex items-center gap-8">
                    <Link href="/products" className="text-sm font-medium text-slate-600 hover:text-primary dark:text-slate-300 dark:hover:text-white transition-colors">
                        Products
                    </Link>
                    <Link href="/solutions" className="text-sm font-medium text-slate-600 hover:text-primary dark:text-slate-300 dark:hover:text-white transition-colors">
                        Solutions
                    </Link>
                    <Link href="/blog" className="text-sm font-medium text-slate-600 hover:text-primary dark:text-slate-300 dark:hover:text-white transition-colors">
                        Blog
                    </Link>
                    <Link href="/pricing" className="text-sm font-medium text-slate-600 hover:text-primary dark:text-slate-300 dark:hover:text-white transition-colors">
                        Pricing
                    </Link>
                </nav>

                {/* Actions */}
                <div className="hidden md:flex items-center gap-4">
                    {/* Theme Toggle */}
                    <button
                        onClick={toggleTheme}
                        className="flex h-9 w-9 items-center justify-center rounded-lg text-slate-500 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-card-dark dark:hover:text-white transition-all"
                        aria-label="Toggle theme"
                    >
                        {mounted ? (
                            theme === "dark" ? (
                                <span className="material-symbols-outlined text-[20px]">light_mode</span>
                            ) : (
                                <span className="material-symbols-outlined text-[20px]">dark_mode</span>
                            )
                        ) : (
                            <span className="w-5 h-5" />
                        )}
                    </button>

                    {/* Auth Buttons */}
                    <div className="flex items-center gap-2">
                        <Link
                            href="/login" // Placeholder URL
                            className="text-sm font-semibold text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white px-4 py-2 transition-colors"
                        >
                            Log in
                        </Link>
                        <Link
                            href="/register"
                            className="inline-flex h-9 items-center justify-center rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-white shadow-sm ring-1 ring-inset ring-primary hover:bg-primary/90 transition-all shadow-primary/25"
                        >
                            Start for free
                        </Link>
                    </div>
                </div>

                {/* Mobile Menu Button */}
                <button
                    className="md:hidden flex h-10 w-10 items-center justify-center rounded-lg text-slate-900 dark:text-white hover:bg-slate-100 dark:hover:bg-card-dark transition-colors"
                    onClick={() => setIsMenuOpen(!isMenuOpen)}
                >
                    <span className="material-symbols-outlined">menu</span>
                </button>
            </div>

            {/* Mobile Menu */}
            {isMenuOpen && (
                <div className="md:hidden border-t border-slate-200 dark:border-border-dark bg-bg-light dark:bg-bg-dark animate-in slide-in-from-top-2">
                    <div className="flex flex-col p-4 space-y-4">
                        <Link href="/products" className="text-base font-medium text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-white">
                            Products
                        </Link>
                        <Link href="/solutions" className="text-base font-medium text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-white">
                            Solutions
                        </Link>
                        <Link href="/blog" className="text-base font-medium text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-white">
                            Blog
                        </Link>
                        <Link href="/pricing" className="text-base font-medium text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-white">
                            Pricing
                        </Link>
                        <div className="h-px bg-slate-200 dark:bg-border-dark my-2" />
                        <Link href="/login" className="flex items-center gap-2 text-base font-medium text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-white">
                            <span className="material-symbols-outlined text-[20px]">login</span> Log in
                        </Link>
                        <Link href="/register" className="flex items-center gap-2 text-base font-medium text-primary hover:text-primary/80">
                            <span className="material-symbols-outlined text-[20px]">rocket_launch</span> Get Started
                        </Link>
                        <div className="flex items-center gap-2 pt-2">
                            <span className="text-sm text-slate-500 dark:text-slate-400">Theme:</span>
                            <button
                                onClick={toggleTheme}
                                className="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 dark:border-border-dark text-slate-500 dark:text-slate-400"
                            >
                                {mounted && (theme === "dark" ? <span className="material-symbols-outlined text-[18px]">light_mode</span> : <span className="material-symbols-outlined text-[18px]">dark_mode</span>)}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </header>
    );
}
