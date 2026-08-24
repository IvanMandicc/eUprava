import { Injectable, signal } from '@angular/core';
import { NotificationService } from './api.services';

// Periodično proverava broj nepročitanih obaveštenja i drži ga u signalu
// da bi navbar mogao da prikaže "oblačić" (bedž) čim stigne nova notifikacija.
const POLL_INTERVAL_MS = 10_000;

@Injectable({ providedIn: 'root' })
export class NotificationBadgeService {
  readonly unreadCount = signal(0);

  private timer?: ReturnType<typeof setInterval>;

  constructor(private notifications: NotificationService) {}

  start(): void {
    if (this.timer) return; // već pokrenuto
    this.refresh();
    this.timer = setInterval(() => this.refresh(), POLL_INTERVAL_MS);
  }

  stop(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = undefined;
    }
    this.unreadCount.set(0);
  }

  refresh(): void {
    this.notifications.list().subscribe({
      next: (list) => this.unreadCount.set(list.filter((n) => !n.read).length),
      error: () => {
        /* tiho ignorišemo — sledeći ciklus će probati ponovo */
      },
    });
  }
}
