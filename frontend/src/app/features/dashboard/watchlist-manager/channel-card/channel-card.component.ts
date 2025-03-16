import {
  ChangeDetectionStrategy,
  Component,
  EventEmitter,
  Input,
  Output,
  inject,
} from "@angular/core";
import { CommonModule } from "@angular/common";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import {
  HlmCardContentDirective,
  HlmCardDirective,
  HlmCardFooterDirective,
  HlmCardHeaderDirective,
  HlmCardTitleDirective,
} from "@spartan-ng/ui-card-helm";
import { LucideAngularModule, Trash2, ExternalLink } from "lucide-angular";
import { Channel } from "../../../../core/models";
import { ChannelService } from "../../../../core/services";
import { ConfirmationDialogComponent } from "../../../../shared/components";

@Component({
  selector: "app-channel-card",
  standalone: true,
  imports: [
    CommonModule,
    HlmButtonDirective,
    HlmCardDirective,
    HlmCardHeaderDirective,
    HlmCardTitleDirective,
    HlmCardContentDirective,
    HlmCardFooterDirective,
    LucideAngularModule,
    ConfirmationDialogComponent,
  ],
  templateUrl: "./channel-card.component.html",
  styleUrls: ["./channel-card.component.css"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ChannelCardComponent {
  @Input({ required: true }) channel!: Channel;
  @Input({ required: true }) watchlistId!: number;
  @Output() channelRemoved = new EventEmitter<void>();

  readonly trashIcon = Trash2;
  readonly externalLinkIcon = ExternalLink;

  private channelService = inject(ChannelService);

  getChannelUrl(): string {
    return `https://www.youtube.com/channel/${this.channel.youtube_id}`;
  }

  getDisplayName(): string {
    return this.channel.custom_name || this.channel.title;
  }

  getTruncatedDescription(): string {
    if (!this.channel.description) return "";
    return this.channel.description.length > 100
      ? `${this.channel.description.substring(0, 100)}...`
      : this.channel.description;
  }

  removeChannel(): void {
    this.channelService
      .removeChannelFromWatchlist(
        this.channel.id,
        String(this.watchlistId),
        this.channel.youtube_id
      )
      .subscribe({
        next: () => {
          this.channelRemoved.emit();
        },
      });
  }
}
