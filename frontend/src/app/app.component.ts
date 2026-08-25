import { Component, effect } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { AuthService } from './core/auth.service';
import { NotificationBadgeService } from './core/notification-badge.service';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  templateUrl: './app.component.html',
})
export class AppComponent {
  constructor(public auth: AuthService, public badge: NotificationBadgeService) {
    // Obaveštenja postoje samo za građane — poluj broj nepročitanih
    // dok god je građanin ulogovan, prestani čim se odjavi ili je druga uloga.
    effect(() => {
      if (this.auth.isCitizen()) {
        this.badge.start();
      } else {
        this.badge.stop();
      }
    });
  }
}
