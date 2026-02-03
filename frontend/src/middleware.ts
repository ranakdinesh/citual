import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
    // Protect /dashboard route
    if (request.nextUrl.pathname.startsWith("/dashboard")) {
        const sessionCookie = request.cookies.get("citual_session");

        if (!sessionCookie) {
            return NextResponse.redirect(new URL("/login", request.url));
        }
    }

    return NextResponse.next();
}

export const config = {
    matcher: [
        "/dashboard/:path*",
    ],
};
