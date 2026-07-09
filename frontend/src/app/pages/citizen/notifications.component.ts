import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { NotificationService } from '../../core/api.services';
import { AppNotification } from '../../models';

@Component({
  selector: 'app-notifications',
  standalone: true,
  imports: [CommonModule],
  template: `
    <h2>Obaveštenja</h2>
    <div class="card">
      @if (notifications.length === 0) {
        <p class="hint">Nemate obaveštenja.</p>
      }
      @for (n of notifications; track n.id) {
        <div class="notification" [class.unread]="!n.read">
          <div>
            <p>{{ n.message }}</p>
            <span class="hint">{{ n.createdAt | date: 'dd.MM.yyyy. HH:mm' }}</span>
          </div>
          @if (!n.read) {
            <button class="secondary" (click)="markRead(n)">Pročitano</button>
          }
        </div>
      }
    </div>
  `,
})
export class NotificationsComponent implements OnInit {
  notifications: AppNotification[] = [];

  constructor(private service: NotificationService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.service.list().subscribe((n) => (this.notifications = n));
  }

  markRead(n: AppNotification): void {
    this.service.markRead(n.id).subscribe(() => (n.read = true));
  }
}
