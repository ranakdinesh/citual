import Link from "next/link";
import { Header } from "@/components/layout/Header";
import { Footer } from "@/components/layout/Footer";

const BLOG_POSTS = [
    {
        id: 1,
        title: "The Rise of Modular Monoliths over Microservices",
        excerpt: "Why successful startups are swtiching back to modular monoliths for better developer velocity and simpler operations.",
        date: "Feb 02, 2026",
        author: "Dinesh",
        category: "Architecture",
        color: "bg-citual-orange",
    },
    {
        id: 2,
        title: "Understanding Hexagonal Architecture in Go",
        excerpt: "A deep dive into Ports and Adapters pattern and how we use it to isolate our domain logic from infrastructure.",
        date: "Jan 28, 2026",
        author: "Team Citual",
        category: "Engineering",
        color: "bg-citual-green",
    },
    {
        id: 3,
        title: "Building Scalable SaaS Multi-tenancy",
        excerpt: "Best practices for implementing strict tenant isolation using Row Level Security (RLS) in PostgreSQL.",
        date: "Jan 15, 2026",
        author: "Engineering",
        category: "Database",
        color: "bg-blue-500",
    },
];

export default function BlogPage() {
    return (
        <>
            <div className="bg-bg-light dark:bg-bg-dark min-h-screen py-24 sm:py-32">
                <div className="mx-auto max-w-7xl px-6 lg:px-8">
                    <div className="mx-auto max-w-2xl text-center">
                        <h2 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white sm:text-4xl">From the Blog</h2>
                        <p className="mt-2 text-lg leading-8 text-slate-600 dark:text-slate-400">
                            Insights on architecture, engineering, and building scalable SaaS products.
                        </p>
                    </div>
                    <div className="mx-auto mt-16 grid max-w-2xl grid-cols-1 gap-x-8 gap-y-20 lg:mx-0 lg:max-w-none lg:grid-cols-3">
                        {BLOG_POSTS.map((post) => (
                            <article key={post.id} className="flex flex-col items-start justify-between p-6 rounded-2xl bg-white dark:bg-card-dark border border-slate-200 dark:border-border-dark shadow-sm hover:shadow-md transition-all hover:border-primary/50">
                                <div className="relative w-full">
                                    <div className={`aspect-[16/9] w-full rounded-2xl ${post.color}/20 mb-5 flex items-center justify-center`}>
                                        <span className="material-symbols-outlined text-4xl text-slate-400/50">article</span>
                                    </div>
                                </div>
                                <div className="flex items-center gap-x-4 text-xs">
                                    <time dateTime={post.date} className="text-slate-500 dark:text-slate-400">
                                        {post.date}
                                    </time>
                                    <span
                                        className="relative z-10 rounded-full bg-slate-100 dark:bg-slate-800 px-3 py-1.5 font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                                    >
                                        {post.category}
                                    </span>
                                </div>
                                <div className="group relative">
                                    <h3 className="mt-3 text-lg font-semibold leading-6 text-slate-900 dark:text-white group-hover:text-primary transition-colors">
                                        <Link href={`/blog/${post.id}`}>
                                            <span className="absolute inset-0" />
                                            {post.title}
                                        </Link>
                                    </h3>
                                    <p className="mt-5 line-clamp-3 text-sm leading-6 text-slate-600 dark:text-slate-400">
                                        {post.excerpt}
                                    </p>
                                </div>
                                <div className="relative mt-8 flex items-center gap-x-4">
                                    <div className="h-10 w-10 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center">
                                        <span className="material-symbols-outlined text-slate-400">person</span>
                                    </div>
                                    <div className="text-sm leading-6">
                                        <p className="font-semibold text-slate-900 dark:text-white">
                                            <a href="#">
                                                <span className="absolute inset-0" />
                                                {post.author}
                                            </a>
                                        </p>
                                    </div>
                                </div>
                            </article>
                        ))}
                    </div>
                </div>
            </div>
        </>
    );
}
