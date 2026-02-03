import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function GET(req: NextRequest) {
    const searchParams = req.nextUrl.searchParams;
    const code = searchParams.get("code");
    const error = searchParams.get("error");

    if (error) {
        return NextResponse.json({ error }, { status: 400 });
    }

    if (!code) {
        return NextResponse.json({ error: "Missing code" }, { status: 400 });
    }

    const backendUrl = process.env.BACKEND_URL || "http://localhost:8090";
    const clientId = process.env.AUTH_CLIENT_ID;
    const clientSecret = process.env.AUTH_CLIENT_SECRET;
    const redirectUri = `${process.env.NEXT_PUBLIC_APP_URL}/api/auth/callback`;

    if (!clientId || !clientSecret) {
        console.error("Missing Client ID or Secret env vars");
        return NextResponse.json({ error: "Server misconfiguration" }, { status: 500 });
    }

    try {
        // Exchange Code for Token
        // We must use Basic Auth for Client Authentication as per OIDC spec for confidential clients
        // Or post body parameters if supported. Fosite supports both. 
        // Let's use Basic Auth header as it is standard and safe.

        // Note: Fosite expects form-urlencoded body
        const params = new URLSearchParams();
        params.append("grant_type", "authorization_code");
        params.append("code", code);
        params.append("redirect_uri", redirectUri);

        const authHeader = `Basic ${Buffer.from(`${encodeURIComponent(clientId)}:${encodeURIComponent(clientSecret)}`).toString("base64")}`;

        const tokenRes = await fetch(`${backendUrl}/oauth2/token`, {
            method: "POST",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded",
                "Authorization": authHeader,
            },
            body: params.toString(),
            cache: "no-store",
        });

        if (!tokenRes.ok) {
            const errorText = await tokenRes.text();
            console.error("Token Exchange Failed:", tokenRes.status, errorText);
            return NextResponse.json({ error: "Failed to exchange token" }, { status: 500 });
        }

        const tokenData = await tokenRes.json();

        // Set Session Cookie (BFF Session)
        // We store the access token (and refresh token/id_token) in an HTTP-only cookie
        // In a real app, you might encrypted this content.
        (await cookies()).set("citual_session", JSON.stringify(tokenData), {
            httpOnly: true,
            secure: process.env.NODE_ENV === "production",
            path: "/",
            sameSite: "lax",
            maxAge: tokenData.expires_in, // Use token expiry
        });

        // Redirect to Dashboard
        return NextResponse.redirect(new URL("/dashboard", req.url));

    } catch (err) {
        console.error("Callback Error:", err);
        return NextResponse.json({ error: "Internal Server Error" }, { status: 500 });
    }
}
