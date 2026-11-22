---
layout: page
title: Wool Privacy Policy
permalink: /wool/privacy/
sitemap: false
---

Effective date: 2026-10-28

Wool is a privacy-first task manager. This policy describes how the iPhone app, and its Mac and Apple Watch companions, handle information for Wool v1.

## Local-first task data

Wool stores tasks, projects, areas, review history, focus sessions, and settings on your device using the local database. Your task data belongs to you.

## Sync

Wool uses Apple CloudKit private database sync when you are signed in to iCloud. Wool app data is stored in your private CloudKit database in the container `iCloud.com.sparkage.wool`, under your own iCloud account. The developer does not operate a custom backend for task data and has no access to your private database.

### Encryption and Advanced Data Protection

CloudKit encrypts private database data in transit and at rest. Wool does not use CloudKit encrypted fields (`encryptedValues`) or any client-side keys of its own, so Wool task data is not end-to-end encrypted. Apple holds the encryption keys for this data, and turning on iCloud Advanced Data Protection does not change that for Wool task data. If you need data that Apple cannot decrypt, keep Wool offline (see "Your choices").

## On-device triage and AI

Wool suggests a status, Area, and due date for captured items. Where Apple Intelligence is available (iOS 26 or later on supported devices), this uses Apple's on-device Foundation Models. Otherwise Wool uses a simple rule-based fallback on your device. Daily briefing text and weekly review prompts are generated the same way. Your task text is not sent to the developer or to any third-party AI service for these features. Suggestions can be wrong, so you can always change them yourself.

## Remote configuration

Wool reads release settings (feature flags) from the public CloudKit database of the Wool container. This request does not include your tasks or other personal data.

## Calendar

If you enable calendar access, Wool reads calendar events to prepare the daily briefing. Calendar data is used locally for read-only display and planning. It is not sent to the developer.

## Voice capture

English voice capture uses Apple's Speech framework. Wool does not require on-device recognition, so depending on your device and settings Apple may process the audio on its servers under Apple's own privacy terms. The developer does not receive or store recordings or transcriptions on a server.

## Sign in with Apple

Sign in with Apple is optional. If you use it, Apple provides the app with an account identifier and, if you share them, your name and email. Wool stores these only in the device Keychain. CloudKit sync depends on your iCloud sign-in, not on Sign in with Apple. Wool does not use Sign in with Apple for tracking.

## Optional integrations not in v1

Gmail import, GitHub Projects mapping, Apple Reminders migration, and Malayalam voice capture are not part of the v1 App Store release. If enabled in a future version, they remain opt-in and require explicit permission.

## Data collection

Wool v1 does not collect personal data on a developer-operated server. Analytics, if present in older code, is not the operational data path for v1. The app contains no advertising SDK and no cross-app tracking.

## Your choices

You can leave your data on device by staying offline or not signing in to iCloud. Sign out leaves local task data on your device. You can clear focus history and local session records in Settings.

## Contact

Questions about this policy can be filed at the Wool issue tracker: https://github.com/kayis-rahman/wool/issues
