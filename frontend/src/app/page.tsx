import Link from "next/link";

export default function Home() {
  return (
    <div className="relative isolate overflow-hidden">
      {/* Background Blobs */}
      <div className="absolute top-0 left-0 w-full h-full overflow-hidden -z-10 pointer-events-none">
        <div className="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-citual-green/20 rounded-full blur-[120px] mix-blend-multiply dark:mix-blend-screen opacity-40 animate-blob"></div>
        <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-citual-orange/15 rounded-full blur-[120px] mix-blend-multiply dark:mix-blend-screen opacity-40 animate-blob animation-delay-2000"></div>
        <div className="absolute top-[20%] left-[20%] w-[50%] h-[50%] bg-primary/10 rounded-full blur-[120px] mix-blend-multiply dark:mix-blend-screen opacity-30 animate-blob animation-delay-4000 hidden md:block"></div>
      </div>

      <div className="px-6 py-24 sm:px-6 sm:py-32 lg:px-8">
        <div className="mx-auto max-w-3xl text-center">
          <div className="inline-flex items-center rounded-full bg-primary/10 px-3 py-1 text-sm font-semibold text-primary ring-1 ring-inset ring-primary/10 mb-6">
            <span className="flex h-2 w-2 rounded-full bg-primary mr-2 animate-pulse"></span>
            The Future of Modular SaaS
          </div>
          <h2 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white sm:text-6xl">
            Scale your business with <br className="hidden sm:block" />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-citual-green to-citual-orange">
              Intelligent Modules
            </span>
          </h2>
          <p className="mx-auto mt-6 max-w-xl text-lg leading-8 text-slate-600 dark:text-slate-300">
            Citual provides a unified platform where Identity, CRM, CMS, and Analytics live in harmony.
            Stop stitching together disconnected tools.
          </p>
          <div className="mt-10 flex items-center justify-center gap-x-6">
            <Link
              href="/register"
              className="rounded-xl bg-primary px-5 py-3 text-sm font-semibold text-white shadow-sm hover:bg-primary/90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary transition-all shadow-primary/25"
            >
              Get started
            </Link>
            <Link
              href="/products"
              className="text-sm font-semibold leading-6 text-slate-900 dark:text-white flex items-center gap-1 group"
            >
              Learn more <span className="material-symbols-outlined text-[18px] transition-transform group-hover:translate-x-1">arrow_forward</span>
            </Link>
          </div>
        </div>
      </div>

      {/* Features Grid Stub */}
      <div className="container mx-auto px-6 pb-24">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {[
            { title: "Identity First", icon: "fingerprint", desc: "Secure, role-based access control with multi-tenant support built-in." },
            { title: "Modular CRM", icon: "groups", desc: "Manage leads, contacts, and deals with a fully integrated CRM module." },
            { title: "Headless CMS", icon: "article", desc: "Publish content anywhere with our powerful, API-first content management system." }
          ].map((feature, idx) => (
            <div key={idx} className="p-8 rounded-2xl bg-white/50 dark:bg-card-dark/50 border border-slate-200 dark:border-border-dark backdrop-blur-sm hover:border-primary/50 transition-colors">
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary mb-6">
                <span className="material-symbols-outlined text-[24px]">{feature.icon}</span>
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3">{feature.title}</h3>
              <p className="text-slate-600 dark:text-slate-400 leading-relaxed">{feature.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
