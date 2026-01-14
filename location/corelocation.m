#import <CoreLocation/CoreLocation.h>
#import <Foundation/Foundation.h>

@interface LocationDelegate : NSObject <CLLocationManagerDelegate>
@property (nonatomic, strong) CLLocationManager *locationManager;
@property (nonatomic) double latitude;
@property (nonatomic) double longitude;
@property (nonatomic) BOOL locationReceived;
@property (nonatomic, strong) NSError *locationError;
@end

@implementation LocationDelegate

- (instancetype)init {
    self = [super init];
    if (self) {
        self.locationManager = [[CLLocationManager alloc] init];
        self.locationManager.delegate = self;
        self.locationManager.desiredAccuracy = kCLLocationAccuracyBest;
        self.locationReceived = NO;
    }
    return self;
}

- (void)locationManager:(CLLocationManager *)manager didUpdateLocations:(NSArray<CLLocation *> *)locations {
    CLLocation *location = [locations lastObject];
    self.latitude = location.coordinate.latitude;
    self.longitude = location.coordinate.longitude;
    self.locationReceived = YES;
    [self.locationManager stopUpdatingLocation];
}

- (void)locationManager:(CLLocationManager *)manager didFailWithError:(NSError *)error {
    self.locationError = error;
    self.locationReceived = YES;
    [self.locationManager stopUpdatingLocation];
}

@end

static LocationDelegate *locationDelegate = nil;

void initLocationManager() {
    if (locationDelegate == nil) {
        locationDelegate = [[LocationDelegate alloc] init];
    }
}

int requestLocation() {
    if (locationDelegate == nil) {
        return -1;
    }

    CLAuthorizationStatus status = [CLLocationManager authorizationStatus];

    // macOS uses kCLAuthorizationStatusAuthorized instead of iOS-specific statuses
    if (status != kCLAuthorizationStatusAuthorized) {
        return -2; // Permission denied or not determined
    }

    if (![CLLocationManager locationServicesEnabled]) {
        return -3; // Location services disabled
    }

    locationDelegate.locationReceived = NO;
    locationDelegate.locationError = nil;

    [locationDelegate.locationManager startUpdatingLocation];

    // Wait for location (up to 10 seconds)
    for (int i = 0; i < 100; i++) {
        [[NSRunLoop currentRunLoop] runUntilDate:[NSDate dateWithTimeIntervalSinceNow:0.1]];
        if (locationDelegate.locationReceived) {
            break;
        }
    }

    if (locationDelegate.locationError != nil) {
        return -4; // Location error
    }

    if (!locationDelegate.locationReceived) {
        return -5; // Timeout
    }

    return 0; // Success
}

double getLatitude() {
    return locationDelegate ? locationDelegate.latitude : 0.0;
}

double getLongitude() {
    return locationDelegate ? locationDelegate.longitude : 0.0;
}