"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import AuthChecker from "@/api/auth/authchecker";

export default function LoginRedirect() {
    const router = useRouter();

    useEffect(() => {
        AuthChecker().then((r) => {
            if (r) router.replace("/");
        });
    }, [router]);

    return null;
}
