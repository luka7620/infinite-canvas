"use client";

import { KeyOutlined, LockOutlined, UserOutlined } from "@ant-design/icons";
import { App, Button, Form, Input, Segmented, Space } from "antd";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

import { fetchCurrentUser } from "@/services/api/auth";
import { useConfigStore } from "@/stores/use-config-store";
import { useUserStore } from "@/stores/use-user-store";

type LoginFormValues = {
    username: string;
    password: string;
    confirmPassword?: string;
    code?: string;
};

export default function LoginPage() {
    return (
        <Suspense fallback={null}>
            <LoginContent />
        </Suspense>
    );
}

function LoginContent() {
    const { message } = App.useApp();
    const router = useRouter();
    const searchParams = useSearchParams();
    const [form] = Form.useForm<LoginFormValues>();
    const login = useUserStore((state) => state.login);
    const register = useUserStore((state) => state.register);
    const setSession = useUserStore((state) => state.setSession);
    const isLoading = useUserStore((state) => state.isLoading);
    const linuxDoEnabled = useConfigStore((state) => state.publicSettings?.auth?.linuxDo?.enabled === true);
    const allowRegister = useConfigStore((state) => state.publicSettings?.auth?.allowRegister !== false);
    const [mode, setMode] = useState<"login" | "register">("login");
    const redirect = searchParams.get("redirect") || "/";
    const inviteCodeRequired = mode === "register" && (!allowRegister || linuxDoEnabled);

    useEffect(() => {
        const token = searchParams.get("token");
        const error = searchParams.get("error");
        if (error) message.error(error);
        if (!token) return;
        void fetchCurrentUser(token).then((user) => {
            setSession(token, user);
            message.success("登录成功");
            router.replace(redirect.startsWith("/") ? redirect : "/");
            router.refresh();
        });
    }, [message, redirect, router, searchParams, setSession]);

    const submit = async (values: LoginFormValues) => {
        try {
            if (mode === "register" && values.password !== values.confirmPassword) {
                message.error("两次输入的密码不一致");
                return;
            }
            const action = mode === "register" ? register : login;
            const user = await action({ username: values.username, password: values.password, code: values.code?.trim() });
            message.success(mode === "register" ? "注册成功" : "登录成功");
            router.replace(redirect.startsWith("/") ? redirect : "/");
            router.refresh();
            if (user.role !== "admin") router.replace("/");
        } catch (error) {
            message.error(error instanceof Error ? error.message : "登录失败");
        }
    };

    const loginWithLinuxDo = async () => {
        try {
            const query = new URLSearchParams({ redirect });
            if (mode === "register") {
                const { code } = await form.validateFields(["code"]);
                const inviteCode = code?.trim();
                if (!inviteCode) {
                    message.error("请输入邀请码");
                    return;
                }
                query.set("code", inviteCode);
            }
            window.location.href = `/api/auth/linux-do/authorize?${query.toString()}`;
        } catch {
            message.error("请输入邀请码");
        }
    };

    return (
        <main className="flex h-full min-h-0 items-center justify-center overflow-y-auto bg-background px-4 py-8 text-foreground">
            <section className="w-full max-w-[420px] rounded-lg border border-border bg-card p-5 sm:p-6">
                <div className="mb-7 text-center">
                    <span
                        className="mx-auto mb-4 block size-12 bg-foreground"
                        style={{
                            mask: "url(/logo.svg) center / contain no-repeat",
                            WebkitMask: "url(/logo.svg) center / contain no-repeat",
                        }}
                        aria-label="无限画布"
                    />
                    <h1 className="text-2xl font-semibold tracking-normal text-foreground">账号登录</h1>
                    <p className="mt-3 text-sm leading-6 text-muted-foreground">支持账号密码和 Linux.do 登录。</p>
                </div>

                <Form<LoginFormValues> form={form} layout="vertical" size="large" requiredMark={false} onFinish={submit}>
                    <Form.Item>
                        <Segmented
                            block
                            value={mode}
                            onChange={(value) => setMode(value as "login" | "register")}
                            options={[{ label: "登录", value: "login" }, { label: "注册", value: "register" }]}
                        />
                    </Form.Item>
                    <Form.Item name="username" label={<span className="font-medium text-foreground">用户名</span>} rules={[{ required: true, message: "请输入用户名" }]}>
                        <Input prefix={<UserOutlined />} autoComplete="username" />
                    </Form.Item>
                    <Form.Item name="password" label={<span className="font-medium text-foreground">密码</span>} rules={[{ required: true, message: "请输入密码" }]}>
                        <Input.Password prefix={<LockOutlined />} autoComplete="current-password" />
                    </Form.Item>
                    {mode === "register" ? (
                        <Form.Item name="confirmPassword" label={<span className="font-medium text-foreground">确认密码</span>} rules={[{ required: true, message: "请再次输入密码" }]}>
                            <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
                        </Form.Item>
                    ) : null}
                    {mode === "register" ? (
                        <Form.Item name="code" label={<span className="font-medium text-foreground">邀请码</span>} rules={inviteCodeRequired ? [{ required: true, message: "请输入邀请码" }] : []}>
                            <Input prefix={<KeyOutlined />} autoComplete="off" />
                        </Form.Item>
                    ) : null}
                    <Space orientation="vertical" size={12} style={{ width: "100%" }}>
                        <Button block type="primary" htmlType="submit" loading={isLoading}>
                            {mode === "register" ? "注册" : "登录"}
                        </Button>
                        {linuxDoEnabled ? (
                            <Button block htmlType="button" onClick={() => void loginWithLinuxDo()} icon={<img src="/icons/linuxdo.svg" alt="" width={18} height={18} />}>
                                使用 Linux.do 登录
                            </Button>
                        ) : null}
                    </Space>
                </Form>
            </section>
        </main>
    );
}
