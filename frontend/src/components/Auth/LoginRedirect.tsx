"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { GetToken } from "@/api/auth/tokenhelpers";

export default function LoginRedirect() {
    const router = useRouter();

    useEffect(() => {
        const token = GetToken();
        if (token && token.trim().length > 0) {
            router.replace("/");
        }
    }, [router]);

    return null;
}