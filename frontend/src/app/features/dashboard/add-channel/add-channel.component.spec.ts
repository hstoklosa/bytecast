import { ComponentFixture, TestBed } from "@angular/core/testing";
import { AddChannelComponent } from "./add-channel.component";
import { WatchlistService } from "../../../core/services/watchlist.service";
import { of } from "rxjs";
import { Channel } from "../../../core/services/watchlist.service";
import { ReactiveFormsModule } from "@angular/forms";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import {
  HlmCardDirective,
  HlmCardContentDirective,
} from "@spartan-ng/ui-card-helm";
import { HlmInputDirective } from "@spartan-ng/ui-input-helm";
import { HlmSpinnerComponent } from "@spartan-ng/ui-spinner-helm";
import { NgIf } from "@angular/common";
import { LucideAngularModule } from "lucide-angular";

describe("AddChannelComponent", () => {
  let component: AddChannelComponent;
  let fixture: ComponentFixture<AddChannelComponent>;
  let mockWatchlistService: jasmine.SpyObj<WatchlistService>;

  beforeEach(async () => {
    mockWatchlistService = jasmine.createSpyObj("WatchlistService", [
      "searchChannels",
      "addChannelToWatchlist",
    ]);

    await TestBed.configureTestingModule({
      imports: [
        ReactiveFormsModule,
        HlmButtonDirective,
        HlmCardDirective,
        HlmCardContentDirective,
        HlmInputDirective,
        HlmSpinnerComponent,
        NgIf,
        LucideAngularModule,
      ],
      declarations: [AddChannelComponent],
      providers: [{ provide: WatchlistService, useValue: mockWatchlistService }],
    }).compileComponents();

    fixture = TestBed.createComponent(AddChannelComponent);
    component = fixture.componentInstance;
    component.watchlistId = 1;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  describe("handleSearch", () => {
    it("should call searchChannels with query and update searchResults", () => {
      const mockResults: Channel[] = [
        {
          id: "123",
          title: "Test Channel",
          description: "Test description",
          thumbnailUrl: "http://test.com",
          subscriberCount: "1000",
        },
      ];

      mockWatchlistService.searchChannels.and.returnValue(of(mockResults));
      component.searchQuery.setValue("test");

      component.handleSearch();

      expect(mockWatchlistService.searchChannels).toHaveBeenCalledWith("test");
      expect(component.searchResults).toEqual(mockResults);
      expect(component.isLoading).toBeFalse();
    });
  });

  describe("addToWatchlist", () => {
    it("should call addChannelToWatchlist and reset form", () => {
      const mockChannel: Channel = {
        id: "123",
        title: "Test Channel",
        description: "Test description",
        thumbnailUrl: "http://test.com",
        subscriberCount: "1000",
      };

      mockWatchlistService.addChannelToWatchlist.and.returnValue(of(void 0));

      component.addToWatchlist(mockChannel);

      expect(mockWatchlistService.addChannelToWatchlist).toHaveBeenCalledWith(
        1,
        "123"
      );
      expect(component.searchResults).toEqual([]);
      expect(component.searchQuery.value).toBeNull();
    });
  });
});
