#!/usr/bin/env python3
"""
Generate synthetic marketing datasets for jq exploration.
Covers a 3-month period with interconnected data.
"""

import json
import random
from datetime import datetime, timedelta
from pathlib import Path
from collections import defaultdict

# Configuration
START_DATE = datetime(2025, 1, 1)
END_DATE = datetime(2025, 3, 31)
BASE_DIR = Path("./data")

# Marketing data fixtures
CAMPAIGNS = [
    "spring_sale_2025", "brand_awareness_q1", "product_launch_mobile",
    "retargeting_winter", "social_media_blitz", "email_nurture_series"
]

PLATFORMS = ["facebook", "google_ads", "instagram", "tiktok", "linkedin", "twitter"]
AD_TYPES = ["video", "carousel", "single_image", "story", "banner"]
DEVICE_TYPES = ["mobile", "desktop", "tablet"]
REGIONS = ["US-WEST", "US-EAST", "EU-CENTRAL", "APAC", "LATAM"]
AUDIENCE_SEGMENTS = ["lookalike", "retargeting", "interest_based", "demographic", "behavioral"]

# Revenue product categories
PRODUCTS = [
    {"id": "PROD-001", "name": "Premium Widget", "price": 299.99},
    {"id": "PROD-002", "name": "Basic Widget", "price": 99.99},
    {"id": "PROD-003", "name": "Widget Pro", "price": 499.99},
    {"id": "PROD-004", "name": "Widget Bundle", "price": 799.99},
    {"id": "PROD-005", "name": "Widget Accessory", "price": 29.99},
]


def generate_ads():
    """Generate ads.json with campaign and ad creative data."""
    ads = []
    ad_id = 1000

    for campaign in CAMPAIGNS:
        num_ads = random.randint(3, 8)
        campaign_budget = random.randint(5000, 50000)

        for i in range(num_ads):
            ad = {
                "ad_id": f"AD-{ad_id}",
                "campaign_id": campaign,
                "campaign_budget": campaign_budget,
                "ad_name": f"{campaign}_creative_{i+1}",
                "platform": random.choice(PLATFORMS),
                "ad_type": random.choice(AD_TYPES),
                "target_audience": {
                    "segments": random.sample(AUDIENCE_SEGMENTS, k=random.randint(1, 3)),
                    "age_range": random.choice(["18-24", "25-34", "35-44", "45-54", "55+"]),
                    "regions": random.sample(REGIONS, k=random.randint(1, 3))
                },
                "creative": {
                    "headline": f"Amazing Deal {i+1}",
                    "cta": random.choice(["Shop Now", "Learn More", "Sign Up", "Get Started"]),
                    "image_url": f"https://cdn.example.com/ad_{ad_id}.jpg"
                },
                "status": random.choice(["active"] * 8 + ["paused", "completed"]),
                "created_at": (START_DATE - timedelta(days=random.randint(7, 30))).isoformat(),
                "daily_budget": round(campaign_budget / 30, 2)
            }
            ads.append(ad)
            ad_id += 1

    return ads


def generate_impressions_for_date(date, ads):
    """Generate impression events for a specific date."""
    impressions = []

    # Generate 500-2000 impressions per day
    num_impressions = random.randint(500, 2000)

    for _ in range(num_impressions):
        ad = random.choice(ads)
        hour = random.randint(0, 23)
        minute = random.randint(0, 59)
        second = random.randint(0, 59)

        timestamp = date.replace(hour=hour, minute=minute, second=second)

        impression = {
            "impression_id": f"IMP-{random.randint(100000, 999999)}",
            "ad_id": ad["ad_id"],
            "campaign_id": ad["campaign_id"],
            "timestamp": timestamp.isoformat(),
            "user_id": f"USER-{random.randint(10000, 99999)}",
            "platform": ad["platform"],
            "device_type": random.choice(DEVICE_TYPES),
            "region": random.choice(ad["target_audience"]["regions"]),
            "clicked": random.random() < 0.08,  # 8% CTR
            "cost_per_impression": round(random.uniform(0.05, 2.5), 3),
            "engagement_time_seconds": random.randint(0, 45) if random.random() < 0.3 else 0
        }
        impressions.append(impression)

    return impressions


def generate_conversions_for_month(year, month, ads):
    """Generate conversion events for a specific month."""
    conversions = []

    # Generate 200-800 conversions per month
    num_conversions = random.randint(200, 800)

    days_in_month = (datetime(year, month % 12 + 1, 1) - timedelta(days=1)).day if month < 12 else 31

    for _ in range(num_conversions):
        ad = random.choice(ads)
        day = random.randint(1, days_in_month)
        hour = random.randint(0, 23)
        minute = random.randint(0, 59)

        timestamp = datetime(year, month, day, hour, minute)

        conversion_type = random.choice(["purchase", "signup", "lead", "download"])

        conversion = {
            "conversion_id": f"CONV-{random.randint(100000, 999999)}",
            "user_id": f"USER-{random.randint(10000, 99999)}",
            "ad_id": ad["ad_id"],
            "campaign_id": ad["campaign_id"],
            "timestamp": timestamp.isoformat(),
            "conversion_type": conversion_type,
            "value": round(random.uniform(10, 1000), 2) if conversion_type == "purchase" else 0,
            "attribution": {
                "model": random.choice(["last_click", "first_click", "linear", "time_decay"]),
                "touchpoints": random.randint(1, 5)
            },
            "device_type": random.choice(DEVICE_TYPES),
            "region": random.choice(REGIONS)
        }
        conversions.append(conversion)

    return conversions


def generate_revenue_for_date(date, ads):
    """Generate revenue transactions for a specific date."""
    revenues = []

    # Generate 20-100 transactions per day
    num_transactions = random.randint(20, 100)

    for _ in range(num_transactions):
        ad = random.choice(ads)
        product = random.choice(PRODUCTS)
        quantity = random.randint(1, 3)
        hour = random.randint(0, 23)
        minute = random.randint(0, 59)

        timestamp = date.replace(hour=hour, minute=minute)

        revenue = {
            "transaction_id": f"TXN-{random.randint(100000, 999999)}",
            "user_id": f"USER-{random.randint(10000, 99999)}",
            "timestamp": timestamp.isoformat(),
            "attributed_ad_id": ad["ad_id"] if random.random() < 0.7 else None,
            "attributed_campaign_id": ad["campaign_id"] if random.random() < 0.7 else None,
            "product": {
                "id": product["id"],
                "name": product["name"],
                "price": product["price"],
                "quantity": quantity
            },
            "total_amount": round(product["price"] * quantity, 2),
            "discount_applied": round(random.uniform(0, 50), 2) if random.random() < 0.2 else 0,
            "payment_method": random.choice(["credit_card", "paypal", "apple_pay", "google_pay"]),
            "region": random.choice(REGIONS),
            "customer_type": random.choice(["new", "returning"])
        }
        revenues.append(revenue)

    return revenues


def main():
    """Generate all marketing datasets."""
    print("Generating marketing datasets...")

    # Create directory structure
    (BASE_DIR / "impressions" / "weekly").mkdir(parents=True, exist_ok=True)
    (BASE_DIR / "impressions" / "monthly" / "2025").mkdir(parents=True, exist_ok=True)
    (BASE_DIR / "conversions").mkdir(parents=True, exist_ok=True)
    (BASE_DIR / "revenue").mkdir(parents=True, exist_ok=True)

    # Generate ads
    print("  Generating ads...")
    ads = generate_ads()
    with open(BASE_DIR / "ads.json", "w") as f:
        json.dump(ads, f, indent=2)

    # Generate impressions (weekly files)
    print("  Generating impressions...")
    current_date = START_DATE
    weekly_impressions = defaultdict(list)
    monthly_impressions = defaultdict(list)

    while current_date <= END_DATE:
        impressions = generate_impressions_for_date(current_date, ads)

        # Weekly file (ISO week start on Monday)
        week_start = current_date - timedelta(days=current_date.weekday())
        weekly_impressions[week_start].extend(impressions)

        # Monthly aggregation
        month_key = f"{current_date.year}-{current_date.month:02d}"
        monthly_impressions[month_key].extend(impressions)

        current_date += timedelta(days=1)

    # Write weekly impression files
    for week_start, impressions in weekly_impressions.items():
        filename = week_start.strftime("%Y-%m-%d.json")
        with open(BASE_DIR / "impressions" / "weekly" / filename, "w") as f:
            json.dump(impressions, f, indent=2)

    # Write monthly impression files
    for month_key, impressions in monthly_impressions.items():
        month_num = month_key.split("-")[1]
        with open(BASE_DIR / "impressions" / "monthly" / "2025" / f"{month_num}.json", "w") as f:
            json.dump(impressions, f, indent=2)

    # Generate conversions (monthly files)
    print("  Generating conversions...")
    for month in range(1, 4):  # Jan, Feb, Mar
        conversions = generate_conversions_for_month(2025, month, ads)
        with open(BASE_DIR / "conversions" / f"2025-{month:02d}.json", "w") as f:
            json.dump(conversions, f, indent=2)

    # Generate revenue (daily files)
    print("  Generating revenue...")
    current_date = START_DATE
    while current_date <= END_DATE:
        revenues = generate_revenue_for_date(current_date, ads)
        filename = current_date.strftime("%Y-%m-%d.json")
        with open(BASE_DIR / "revenue" / filename, "w") as f:
            json.dump(revenues, f, indent=2)
        current_date += timedelta(days=1)

    # Generate summary
    print("\nDataset Generation Complete!")
    print(f"  Ads: {len(ads)}")
    print(f"  Weekly impression files: {len(weekly_impressions)}")
    print(f"  Monthly impression files: 3")
    print(f"  Conversion files: 3")
    print(f"  Revenue files: {(END_DATE - START_DATE).days + 1}")
    print(f"\nData location: {BASE_DIR.absolute()}")


if __name__ == "__main__":
    main()
