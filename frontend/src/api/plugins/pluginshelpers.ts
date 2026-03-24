export type Hook = { friendly_name: string; signature: string; type: string };

export type Hooks = Record<string, Hook>;

export type Plugin = {
    enabled: boolean;
    name: string;
    friendly_name: string;
    version: string;
    hooks: Hooks;
};
