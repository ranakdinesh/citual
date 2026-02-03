"use client";

import Link from "next/link";
import { useState } from "react";
import { useRouter } from "next/navigation";

export default function LoginPage() {
    const router = useRouter();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError("");

        const formData = new FormData();
        formData.append("email", email);
        formData.append("password", password);

        // Call Server Action
        import("../actions/auth").then(async (mod) => {
            const result = await mod.loginAction(formData);

            if (result.error) {
                setError(result.error);
                setLoading(false);
                return;
            }

            if (result.redirectUrl) {
                // Redirect user to the OAuth2 Authorize URL
                window.location.href = result.redirectUrl;
            } else {
                setLoading(false);
            }
        });
    };

    return (
        <div className="relative min-h-screen w-full flex flex-col items-center justify-center p-4 overflow-hidden bg-bg-light dark:bg-bg-dark font-display text-slate-900 dark:text-white antialiased transition-colors duration-300">
            {/* Background Blobs */}
            <div className="absolute top-0 left-0 w-full h-full overflow-hidden -z-10 pointer-events-none">
                <div className="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-citual-green/20 rounded-full blur-[120px] mix-blend-multiply dark:mix-blend-screen opacity-40 animate-blob"></div>
                <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-citual-orange/15 rounded-full blur-[120px] mix-blend-multiply dark:mix-blend-screen opacity-40 animate-blob animation-delay-2000"></div>
            </div>

            <div className="mb-8 flex flex-col items-center z-10">
                <Link href="/" className="flex items-center gap-3 group">
                    <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-white dark:bg-card-dark shadow-xl text-primary group-hover:scale-105 transition-transform">
                        <span className="material-symbols-outlined text-[40px]">dashboard_customize</span>
                    </div>
                </Link>
            </div>

            <div className="w-full max-w-[440px] flex flex-col gap-6 relative z-10">
                <form onSubmit={handleLogin} className="flex flex-col rounded-2xl bg-white dark:bg-card-dark border border-slate-200 dark:border-border-dark shadow-2xl overflow-hidden">
                    <div className="px-8 pt-8 pb-4 text-center">
                        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white mb-2">Welcome to Citual</h1>
                        <p className="text-slate-500 dark:text-slate-400 text-sm">Log in to access your workspace</p>
                    </div>

                    <div className="border-b border-slate-200 dark:border-border-dark">
                        <div className="flex px-4 justify-between">
                            <button type="button" className="group flex flex-col items-center justify-center border-b-[3px] border-b-transparent hover:border-b-slate-300 dark:hover:border-b-slate-600 text-slate-400 dark:text-slate-400 gap-2 pb-3 pt-4 flex-1 transition-colors">
                                <span className="material-symbols-outlined text-[24px] group-hover:text-slate-600 dark:group-hover:text-slate-200">person</span>
                                <span className="text-xs font-semibold uppercase tracking-wider">Username</span>
                            </button>
                            <button type="button" className="flex flex-col items-center justify-center border-b-[3px] border-b-citual-orange text-citual-orange gap-2 pb-3 pt-4 flex-1">
                                <span className="material-symbols-outlined text-[24px] fill-1">mail</span>
                                <span className="text-xs font-semibold uppercase tracking-wider">Email</span>
                            </button>
                            <button type="button" className="group flex flex-col items-center justify-center border-b-[3px] border-b-transparent hover:border-b-slate-300 dark:hover:border-b-slate-600 text-slate-400 dark:text-slate-400 gap-2 pb-3 pt-4 flex-1 transition-colors">
                                <span className="material-symbols-outlined text-[24px] group-hover:text-slate-600 dark:group-hover:text-slate-200">smartphone</span>
                                <span className="text-xs font-semibold uppercase tracking-wider">Mobile</span>
                            </button>
                        </div>
                    </div>

                    <div className="flex flex-col p-6 gap-5">
                        {error && (
                            <div className="p-3 bg-red-50 text-red-600 text-sm rounded-lg border border-red-200 dark:bg-red-900/20 dark:border-red-800 dark:text-red-400">
                                {error}
                            </div>
                        )}

                        <label className="flex flex-col gap-2">
                            <span className="text-slate-700 dark:text-white text-sm font-medium">Email Address</span>
                            <div className="relative group">
                                <span className="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 pointer-events-none group-focus-within:text-citual-green transition-colors">alternate_email</span>
                                <input
                                    className="w-full rounded-xl bg-slate-50 dark:bg-input-dark border border-slate-200 dark:border-border-dark text-slate-900 dark:text-white text-sm placeholder:text-slate-400 dark:placeholder:text-slate-500 h-12 pl-11 pr-4 focus:ring-2 focus:ring-citual-green/20 focus:border-citual-green transition-all outline-none"
                                    placeholder="name@company.com"
                                    type="email"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    required
                                />
                            </div>
                        </label>

                        <div className="flex w-full items-center justify-center rounded-lg bg-slate-100 dark:bg-input-dark p-1">
                            <label className="cursor-pointer relative flex-1 flex items-center justify-center py-1.5 rounded-md transition-all has-[:checked]:bg-white dark:has-[:checked]:bg-[#324467] has-[:checked]:shadow-sm has-[:checked]:text-citual-green dark:has-[:checked]:text-white text-slate-500 dark:text-slate-400">
                                <span className="text-sm font-medium z-10">Password</span>
                                <input defaultChecked className="sr-only" name="auth-method" type="radio" value="password" />
                            </label>
                            <label className="cursor-pointer relative flex-1 flex items-center justify-center py-1.5 rounded-md transition-all has-[:checked]:bg-white dark:has-[:checked]:bg-[#324467] has-[:checked]:shadow-sm has-[:checked]:text-citual-green dark:has-[:checked]:text-white text-slate-500 dark:text-slate-400">
                                <span className="text-sm font-medium z-10">One-Time Password</span>
                                <input className="sr-only" name="auth-method" type="radio" value="otp" />
                            </label>
                        </div>

                        <label className="flex flex-col gap-2">
                            <div className="flex justify-between items-center">
                                <span className="text-slate-700 dark:text-white text-sm font-medium">Password</span>
                                <a href="#" className="text-xs font-medium text-citual-orange hover:text-citual-orange/80 transition-colors">Forgot Password?</a>
                            </div>
                            <div className="relative group">
                                <span className="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 pointer-events-none group-focus-within:text-citual-green transition-colors">lock</span>
                                <input
                                    className="w-full rounded-xl bg-slate-50 dark:bg-input-dark border border-slate-200 dark:border-border-dark text-slate-900 dark:text-white text-sm placeholder:text-slate-400 dark:placeholder:text-slate-500 h-12 pl-11 pr-11 focus:ring-2 focus:ring-citual-green/20 focus:border-citual-green transition-all outline-none"
                                    placeholder="Enter your password"
                                    type="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    required
                                />
                                <button className="absolute right-0 top-0 h-full px-4 flex items-center text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-white transition-colors" type="button">
                                    <span className="material-symbols-outlined text-[20px]">visibility</span>
                                </button>
                            </div>
                        </label>

                        <button
                            type="submit"
                            disabled={loading}
                            className="mt-2 w-full h-12 bg-citual-green hover:bg-[#4a725a] active:bg-[#3d614a] text-white font-semibold rounded-xl transition-all shadow-lg shadow-citual-green/25 flex items-center justify-center gap-2 group disabled:opacity-70 disabled:cursor-not-allowed"
                        >
                            {loading ? (
                                <span>Logging in...</span>
                            ) : (
                                <>
                                    <span>Log In</span>
                                    <span className="material-symbols-outlined text-sm transition-transform group-hover:translate-x-1">arrow_forward</span>
                                </>
                            )}
                        </button>

                        <div className="relative flex py-2 items-center">
                            <div className="flex-grow border-t border-slate-200 dark:border-border-dark"></div>
                            <span className="flex-shrink mx-4 text-slate-400 text-xs font-medium uppercase">Or continue with</span>
                            <div className="flex-grow border-t border-slate-200 dark:border-border-dark"></div>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                            <button type="button" className="flex h-10 items-center justify-center gap-2 rounded-lg border border-slate-200 dark:border-border-dark bg-white dark:bg-input-dark hover:bg-slate-50 dark:hover:bg-[#2b3952] transition-colors">
                                <img alt="Google" className="h-5 w-5" src="https://lh3.googleusercontent.com/aida-public/AB6AXuDJLRmNVPgqSdV51BtejZkqYgBSbsM70NcPiy97mA-jWENp-Rfa1rsxgBnesiAk5k4iuWkoYphSlVMEjGtYxUKR28jnKkhmWuNcHPxe5zSUksbzofmJIAgOkIo23zqmlRPVL48jWoS10sk1j33rbXXyW2ZsPPvzVUbXtp-nZuQBl-ONhuDgU7LwbJY-MYjRBD3jGN_wjMpUWfb4-ysrgAvdUX6rGswKyJZdQjaO3eEOsqcBiKB9t_ZZzONhek9mMDwGXE-BWs9bEvjY" />
                                <span className="text-sm font-medium text-slate-700 dark:text-slate-200">Google</span>
                            </button>
                            <button type="button" className="flex h-10 items-center justify-center gap-2 rounded-lg border border-slate-200 dark:border-border-dark bg-white dark:bg-input-dark hover:bg-slate-50 dark:hover:bg-[#2b3952] transition-colors">
                                <img alt="Microsoft" className="h-5 w-5" src="https://lh3.googleusercontent.com/aida-public/AB6AXuCxueFM5PO5cmaotc_dInf_FpCJaeXgnmq5ap3Q5Z1qefKOtJcEOHJus2NlIG03bg1KE2JwHaPzGPEy3GAN7lT9kwkMUhzkYr3DHr-dnWvnwa8-hfxxfGM-PBRyDjZPO4_zZyo00Ig8UZ4QwpusAP70lUm1MrHNhiTgif0DuHtG3aVyGnvND5RcS97IEDsYIONSD_VAuHPBzCrR_Qt2rveGLGxQk2ChkZQseHdH_EuZg-d-eaqthsINLCuekSBhEAoBjNheWJEfGq7i" />
                                <span className="text-sm font-medium text-slate-700 dark:text-slate-200">Microsoft</span>
                            </button>
                        </div>
                    </div>

                    <div className="bg-slate-50 dark:bg-[#151b28] py-4 px-6 text-center border-t border-slate-200 dark:border-border-dark">
                        <p className="text-sm text-slate-500 dark:text-slate-400">
                            Do not have an account? <a href="/register" className="font-semibold text-citual-orange hover:text-citual-orange/80 transition-colors">Sign up</a>
                        </p>
                    </div>
                </form>

                <div className="flex justify-center gap-6 text-sm text-slate-500 dark:text-slate-400">
                    <a href="#" className="hover:text-citual-green dark:hover:text-slate-200 transition-colors">Privacy Policy</a>
                    <a href="#" className="hover:text-citual-green dark:hover:text-slate-200 transition-colors">Terms of Service</a>
                    <a href="#" className="hover:text-citual-green dark:hover:text-slate-200 transition-colors">Help Center</a>
                </div>
            </div>
        </div>
    );
}
