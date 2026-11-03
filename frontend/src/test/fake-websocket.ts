import { vi } from 'vitest';

/**
 * Minimal in-memory WebSocket double for hook tests. Install with
 * `installFakeWebSocket()`; every constructed socket is recorded in
 * `FakeWebSocket.instances`.
 */
export class FakeWebSocket {
    static readonly CONNECTING = 0;
    static readonly OPEN = 1;
    static readonly CLOSING = 2;
    static readonly CLOSED = 3;
    static instances: FakeWebSocket[] = [];

    readonly url: string;
    readonly protocols: string | string[] | undefined;
    readyState = FakeWebSocket.CONNECTING;
    sent: string[] = [];
    closeCalls = 0;
    onopen: ((event: Event) => void) | null = null;
    onmessage: ((event: MessageEvent) => void) | null = null;
    onerror: ((event: Event) => void) | null = null;
    onclose: ((event: CloseEvent) => void) | null = null;

    constructor(url: string, protocols?: string | string[]) {
        this.url = url;
        this.protocols = protocols;
        FakeWebSocket.instances.push(this);
    }

    send(data: string) {
        this.sent.push(data);
    }

    /**
     * Client-initiated close. Like a browser, the close event is delivered
     * asynchronously to whatever handler is attached at that time.
     */
    close() {
        this.closeCalls += 1;
        if (this.readyState >= FakeWebSocket.CLOSING) return;
        this.readyState = FakeWebSocket.CLOSING;
        queueMicrotask(() => {
            this.readyState = FakeWebSocket.CLOSED;
            this.onclose?.({ code: 1000 } as CloseEvent);
        });
    }

    serverOpen() {
        this.readyState = FakeWebSocket.OPEN;
        this.onopen?.(new Event('open'));
    }

    serverSend(payload: unknown) {
        this.onmessage?.({ data: JSON.stringify(payload) } as MessageEvent);
    }

    /** Server/network-initiated close. */
    serverClose(code = 1006) {
        this.readyState = FakeWebSocket.CLOSED;
        this.onclose?.({ code } as CloseEvent);
    }
}

export function installFakeWebSocket() {
    FakeWebSocket.instances = [];
    vi.stubGlobal('WebSocket', FakeWebSocket);
    return FakeWebSocket;
}
