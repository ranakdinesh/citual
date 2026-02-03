import { cookies } from "next/headers";
import { redirect } from "next/navigation";

export default async function DashboardPage() {
    const sessionCookie = (await cookies()).get("citual_session");

    if (!sessionCookie) {
        redirect("/login");
    }

    let user = null;
    try {
        const tokenData = JSON.parse(sessionCookie.value);
        // In a real app, you might decode the JWT here to get user info
        user = tokenData;
    } catch (e) {
        // invalid token data
    }

    return (
        <div className="p-8">
            <div className="max-w-4xl mx-auto">
                <h1 className="text-3xl font-bold mb-6 text-slate-900 dark:text-white">Dashboard</h1>

                <div className="bg-white dark:bg-card-dark rounded-xl shadow-sm border border-slate-200 dark:border-border-dark p-6">
                    <h2 className="text-xl font-semibold mb-4 text-citual-green">Welcome back!</h2>
                    <p className="text-slate-600 dark:text-slate-300 mb-4">
                        You are successfully logged in.
                    </p>

                    <div className="bg-slate-100 dark:bg-[#151b28] p-4 rounded-lg overflow-x-auto">
                        <pre className="text-xs text-slate-700 dark:text-slate-300">
                            {JSON.stringify(user, null, 2)}
                        </pre>
                    </div>
                </div>
            </div>
        </div>
    );
}
