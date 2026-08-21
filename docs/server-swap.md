# Swap on the production server

Added 2026-08-21 after ClamAV was OOM-killed twice in two days, the second
time taking the API container down with it for roughly two minutes. The box
has 7.7 GB of RAM and had **no swap at all**, so a single process asking for
more than the kernel could free was killed outright rather than slowed down.

What is in place:

| | |
|---|---|
| File | `/swapfile`, 2 GB, mode 600 |
| Persistence | `/swapfile none swap sw 0 0` in `/etc/fstab` |
| Swappiness | `10`, set in `/etc/sysctl.d/99-swappiness.conf` |

The fstab entry was verified by taking the swap down and bringing it back
with `swapon -a` rather than by reading the line and assuming it parses. A
mistake in fstab is the kind that only shows up at the next reboot, which is
the worst moment to find it.

Swappiness is 10 rather than the default 60 on purpose. Swap here is an
emergency backstop, not a routine tier: the workload fits in RAM, and the
only thing swap has to do is give the kernel somewhere to put pages during a
spike instead of killing whatever asked for memory. At 60 the kernel pushes
pages out proactively, which on a database host trades latency for nothing.

## What this does not fix

ClamAV's signature database is what blows up, during `freshclam` updates. Swap
means the update is slow instead of fatal, and the API no longer dies with it,
because `api` depends on `clamav` being healthy. If it recurs anyway, the next
step is a memory limit on the clamav container so it fails alone.

`/etc/fstab.bak-20260821` holds the previous file.
