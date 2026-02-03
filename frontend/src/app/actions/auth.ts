"use server";

import { cookies } from "next/headers";

export async function loginAction(formData: FormData) {
    const email = formData.get("email");
    const password = formData.get("password");

    if (!email || !password) {
        return { error: "Email and password are required" };
    }

    const backendUrl = process.env.BACKEND_URL || "http://localhost:8090";

    try {
        // 1. Authenticate with IDP (Backend)
        const res = await fetch(`${backendUrl}/auth/login`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ email, password }),
            cache: "no-store",
        });

        if (!res.ok) {
            const data = await res.json().catch(() => ({}));
            return { error: data.message || "Invalid credentials" };
        }

        // 2. Extract SSO Cookie
        const setCookieHeader = res.headers.get("set-cookie");
        if (setCookieHeader) {
            // Simple parsing - assumes standard format
            // In production, might need a more robust parser or multiple headers
            // Robust cookie parsing: "name=value; Path=/; ..."
            const firstPart = setCookieHeader.split(";")[0];
            const separatorIndex = firstPart.indexOf("=");
            const name = firstPart.substring(0, separatorIndex);
            const value = firstPart.substring(separatorIndex + 1);

            (await cookies()).set({
                name,
                value,
                httpOnly: true,
                path: "/",
                sameSite: "lax",
                secure: process.env.NODE_ENV === "production",
            });
        }

        // 3. Construct OIDC Authorization URL
        // This directs the user to the IDP's authorize endpoint, which will now see the cookie we just set
        const clientId = process.env.AUTH_CLIENT_ID;
        const redirectUri = `${process.env.NEXT_PUBLIC_APP_URL}/api/auth/callback`;
        const scopes = "openid profile email offline_access";
        const state = "random_state_string"; // Should be random in prod

        if (!clientId) {
            throw new Error("Missing Client ID configuration");
        }

        const authorizeUrl = `${backendUrl}/oauth2/auth?response_type=code&client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&scope=${encodeURIComponent(scopes)}&state=${state}`;

        // Return the redirect URL to the client component
        return { success: true, redirectUrl: authorizeUrl };

    } catch (error: any) {
        console.error("Login Action Error:", error);
        return { error: "An unexpected error occurred. Please try again." };
    }
}
