//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

extern void skillflowTrayOnShow(void);
extern void skillflowTrayOnHide(void);
extern void skillflowTrayOnQuit(void);
extern void skillflowTrayOnApplicationWillHide(void);
extern void skillflowTrayOnApplicationDidHide(void);
extern void skillflowTrayLog(const char *level, const char *message);

@interface SkillFlowTrayDelegate : NSObject
@end

static NSStatusItem *skillflowStatusItem = nil;
static NSMenu *skillflowStatusMenu = nil;
static SkillFlowTrayDelegate *skillflowTrayDelegate = nil;
static NSImage *skillflowStatusImage = nil;
static BOOL skillflowTrayObserversRegistered = NO;

static void skillflow_log(NSString *level, NSString *message) {
	skillflowTrayLog([level UTF8String], [message UTF8String]);
}

static void skillflow_log_debug(NSString *message) {
	skillflow_log(@"debug", message);
}

static void skillflow_log_info(NSString *message) {
	skillflow_log(@"info", message);
}

static void skillflow_log_error(NSString *message) {
	skillflow_log(@"error", message);
}

static void skillflow_run_on_main_thread_sync(dispatch_block_t block) {
	if ([NSThread isMainThread]) {
		block();
		return;
	}
	dispatch_sync(dispatch_get_main_queue(), block);
}

static void skillflow_run_on_main_thread_async(dispatch_block_t block) {
	dispatch_async(dispatch_get_main_queue(), block);
}

void skillflow_prepare_application(void) {
	skillflow_run_on_main_thread_sync(^{
		[NSApplication sharedApplication];
		[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
	});
}

void skillflow_run_application(void) {
	[NSApp run];
}

void skillflow_stop_application(void) {
	skillflow_run_on_main_thread_async(^{
		[NSApp terminate:nil];
	});
}

static NSImage *skillflow_make_template_status_image(void) {
	NSSize size = NSMakeSize(18, 18);
	NSImage *image = [[NSImage alloc] initWithSize:size];
	[image lockFocus];
	[[NSGraphicsContext currentContext] setShouldAntialias:YES];
	[[NSColor blackColor] setStroke];
	[[NSColor blackColor] setFill];

	NSBezierPath *flow = [NSBezierPath bezierPath];
	[flow setLineWidth:1.9];
	[flow setLineCapStyle:NSLineCapStyleRound];
	[flow setLineJoinStyle:NSLineJoinStyleRound];
	[flow moveToPoint:NSMakePoint(4.2, 13.2)];
	[flow curveToPoint:NSMakePoint(13.6, 13.0)
		controlPoint1:NSMakePoint(4.8, 16.0)
		controlPoint2:NSMakePoint(12.7, 16.0)];
	[flow curveToPoint:NSMakePoint(9.4, 9.2)
		controlPoint1:NSMakePoint(14.8, 10.9)
		controlPoint2:NSMakePoint(11.8, 10.6)];
	[flow curveToPoint:NSMakePoint(4.5, 5.4)
		controlPoint1:NSMakePoint(6.8, 8.0)
		controlPoint2:NSMakePoint(4.8, 7.0)];
	[flow curveToPoint:NSMakePoint(13.4, 5.0)
		controlPoint1:NSMakePoint(3.7, 2.0)
		controlPoint2:NSMakePoint(12.4, 1.8)];
	[flow stroke];

	NSBezierPath *headTop = [NSBezierPath bezierPath];
	[headTop moveToPoint:NSMakePoint(12.2, 14.4)];
	[headTop lineToPoint:NSMakePoint(14.9, 13.0)];
	[headTop lineToPoint:NSMakePoint(12.2, 11.5)];
	[headTop closePath];
	[headTop fill];

	NSBezierPath *headBottom = [NSBezierPath bezierPath];
	[headBottom moveToPoint:NSMakePoint(5.8, 6.8)];
	[headBottom lineToPoint:NSMakePoint(3.1, 5.4)];
	[headBottom lineToPoint:NSMakePoint(5.8, 3.9)];
	[headBottom closePath];
	[headBottom fill];

	[image unlockFocus];
	[image setTemplate:YES];
	[image setSize:size];
	return image;
}

static void skillflow_build_status_menu(void) {
	if (skillflowStatusMenu != nil) {
		return;
	}

	skillflowStatusMenu = [[NSMenu alloc] initWithTitle:@"SkillFlow"];

	NSMenuItem *showItem = [[NSMenuItem alloc] initWithTitle:@"Show SkillFlow" action:@selector(onShow:) keyEquivalent:@""];
	[showItem setTarget:skillflowTrayDelegate];
	[skillflowStatusMenu addItem:showItem];

	NSMenuItem *hideItem = [[NSMenuItem alloc] initWithTitle:@"Hide SkillFlow" action:@selector(onHide:) keyEquivalent:@""];
	[hideItem setTarget:skillflowTrayDelegate];
	[skillflowStatusMenu addItem:hideItem];

	[skillflowStatusMenu addItem:[NSMenuItem separatorItem]];

	NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"Quit SkillFlow" action:@selector(onQuit:) keyEquivalent:@""];
	[quitItem setTarget:skillflowTrayDelegate];
	[skillflowStatusMenu addItem:quitItem];
}

static BOOL skillflow_configure_status_item_button(void) {
	if (skillflowStatusItem == nil) {
		skillflow_log_error(@"tray status item configure failed: status item missing");
		return NO;
	}

	NSStatusBarButton *button = [skillflowStatusItem button];
	if (button == nil) {
		skillflow_log_error(@"tray status item configure failed: status button missing");
		return NO;
	}

	if (skillflowStatusImage == nil) {
		skillflowStatusImage = skillflow_make_template_status_image();
	}
	if (skillflowStatusImage == nil) {
		skillflow_log_error(@"tray status item configure failed: icon generation returned nil");
		return NO;
	}

	[button setImage:skillflowStatusImage];
	[button setImagePosition:NSImageOnly];
	[button setToolTip:@"SkillFlow"];
		[skillflowStatusItem setMenu:skillflowStatusMenu];
	if ([skillflowStatusItem respondsToSelector:@selector(setVisible:)]) {
		[(id)skillflowStatusItem setVisible:YES];
	}
	if ([button respondsToSelector:@selector(setImageScaling:)]) {
		[button setImageScaling:NSImageScaleProportionallyDown];
	}
	[button setNeedsDisplay:YES];
	NSRect frame = [button frame];
	BOOL visible = YES;
	if ([skillflowStatusItem respondsToSelector:@selector(isVisible)]) {
		visible = [(id)skillflowStatusItem isVisible];
	}
	skillflow_log_debug([NSString stringWithFormat:@"tray status item configured, icon=template visible=%@ width=%.1f height=%.1f", visible ? @"true" : @"false", frame.size.width, frame.size.height]);
	return YES;
}

static BOOL skillflow_create_status_item(void) {
	if (skillflowTrayDelegate == nil) {
		skillflowTrayDelegate = [[SkillFlowTrayDelegate alloc] init];
		if (skillflowTrayDelegate == nil) {
			skillflow_log_error(@"tray status item create failed: delegate init failed");
			return NO;
		}
		skillflow_log_debug(@"tray delegate created");
	}

	skillflow_build_status_menu();

	if (skillflowStatusItem == nil) {
		skillflowStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSSquareStatusItemLength];
		if (skillflowStatusItem == nil) {
			skillflow_log_error(@"tray status item create failed: NSStatusBar returned nil");
			return NO;
		}
		[skillflowStatusItem retain];
			skillflow_log_debug([NSString stringWithFormat:@"tray status item created, ptr=%p", skillflowStatusItem]);
	} else {
		skillflow_log_debug([NSString stringWithFormat:@"tray status item reused, ptr=%p", skillflowStatusItem]);
	}

	return skillflow_configure_status_item_button();
}

static void skillflow_schedule_status_item_retry(NSTimeInterval delaySeconds) {
	dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(delaySeconds * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
		if (skillflowStatusItem != nil) {
			return;
		}
		skillflow_log_debug([NSString stringWithFormat:@"tray status item retry started, delay=%.2fs", delaySeconds]);
		if (skillflow_create_status_item()) {
			skillflow_log_debug(@"tray status item retry completed");
			return;
		}
		skillflow_log_error(@"tray status item retry failed");
	});
}

@implementation SkillFlowTrayDelegate
- (void)onShow:(id)sender {
	skillflowTrayOnShow();
}

- (void)onHide:(id)sender {
	skillflowTrayOnHide();
}

- (void)onQuit:(id)sender {
	skillflowTrayOnQuit();
}

- (void)onApplicationWillHide:(NSNotification *)notification {
	skillflowTrayOnApplicationWillHide();
}

- (void)onApplicationDidHide:(NSNotification *)notification {
	skillflowTrayOnApplicationDidHide();
}

- (void)onApplicationDidFinishLaunching:(NSNotification *)notification {
	skillflow_log_debug(@"tray observed application did finish launching");
	if (!skillflow_create_status_item()) {
		skillflow_schedule_status_item_retry(0.25);
		skillflow_schedule_status_item_retry(1.00);
	}
}
@end

static void skillflow_register_tray_notifications(void) {
	if (skillflowTrayObserversRegistered || skillflowTrayDelegate == nil) {
		return;
	}
	[[NSNotificationCenter defaultCenter] addObserver:skillflowTrayDelegate selector:@selector(onApplicationWillHide:) name:NSApplicationWillHideNotification object:NSApp];
	[[NSNotificationCenter defaultCenter] addObserver:skillflowTrayDelegate selector:@selector(onApplicationDidHide:) name:NSApplicationDidHideNotification object:NSApp];
	[[NSNotificationCenter defaultCenter] addObserver:skillflowTrayDelegate selector:@selector(onApplicationDidFinishLaunching:) name:NSApplicationDidFinishLaunchingNotification object:NSApp];
	skillflowTrayObserversRegistered = YES;
	skillflow_log_debug(@"tray notifications registered");
}

static int skillflow_setup_tray(void) {
	skillflow_run_on_main_thread_async(^{
		skillflow_log_debug(@"tray native setup started");
		if (skillflowTrayDelegate == nil) {
			skillflowTrayDelegate = [[SkillFlowTrayDelegate alloc] init];
			if (skillflowTrayDelegate == nil) {
				skillflow_log_error(@"tray native setup failed: delegate init failed");
				return;
			}
		}

		skillflow_register_tray_notifications();
		if (!skillflow_create_status_item()) {
			skillflow_schedule_status_item_retry(0.25);
			skillflow_schedule_status_item_retry(1.00);
			return;
		}
		skillflow_log_info(@"tray native setup completed");
	});
	return 1;
}

static void skillflow_apply_accessory_policy(void) {
	skillflow_run_on_main_thread_sync(^{
		[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
		skillflow_log_debug(@"tray activation policy set, mode=accessory");
	});
}

static void skillflow_apply_regular_policy(void) {
	skillflow_run_on_main_thread_sync(^{
		skillflow_create_status_item();
		[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
		skillflow_log_debug(@"tray activation policy set, mode=regular");
	});
}

static void skillflow_ensure_status_item(void) {
	skillflow_run_on_main_thread_sync(^{
		skillflow_create_status_item();
		skillflow_log_debug(@"tray status item ensured");
	});
}

static void skillflow_teardown_tray(void) {
	skillflow_run_on_main_thread_sync(^{
		if (skillflowTrayObserversRegistered && skillflowTrayDelegate != nil) {
			[[NSNotificationCenter defaultCenter] removeObserver:skillflowTrayDelegate];
			skillflowTrayObserversRegistered = NO;
		}
		if (skillflowStatusItem != nil) {
			[[NSStatusBar systemStatusBar] removeStatusItem:skillflowStatusItem];
			[skillflowStatusItem release];
			skillflowStatusItem = nil;
		}
		if (skillflowStatusImage != nil) {
			[skillflowStatusImage release];
			skillflowStatusImage = nil;
		}
		skillflowStatusMenu = nil;
		skillflowTrayDelegate = nil;
		skillflow_log_debug(@"tray native teardown completed");
	});
}
*/
import "C"

import "sync"

var darwinTrayState struct {
	mu         sync.RWMutex
	controller trayController
}

func setupTray(controller trayController) error {
	darwinTrayState.mu.Lock()
	darwinTrayState.controller = controller
	darwinTrayState.mu.Unlock()

	controller.logInfof("tray setup started, platform=darwin")
	if C.skillflow_setup_tray() == 0 {
		controller.logErrorf("tray setup failed: create menu bar status item failed")
		return errDarwinTraySetup
	}
	controller.logDebugf("tray setup queued, platform=darwin")
	return nil
}

func teardownTray() {
	darwinTrayState.mu.Lock()
	controller := darwinTrayState.controller
	darwinTrayState.controller = nil
	darwinTrayState.mu.Unlock()
	if controller != nil {
		controller.logDebugf("tray teardown started, platform=darwin")
	}
	C.skillflow_teardown_tray()
	if controller != nil {
		controller.logDebugf("tray teardown completed, platform=darwin")
	}
}

func withDarwinTrayController(fn func(trayController)) {
	darwinTrayState.mu.RLock()
	controller := darwinTrayState.controller
	darwinTrayState.mu.RUnlock()
	if controller != nil {
		fn(controller)
	}
}

func applyDarwinAccessoryPolicy() {
	C.skillflow_apply_accessory_policy()
}

func applyDarwinRegularPolicy() {
	C.skillflow_apply_regular_policy()
}

func ensureDarwinStatusItem() {
	C.skillflow_ensure_status_item()
}
