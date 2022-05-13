#!/usr/bin/env bash
##############################################################################
# Build a macOS installer package containing a weep universal binary
# Author: Patrick Sanders <psanders@netflix.com>
##############################################################################
set -euo pipefail

BASE_DIR="build/package/macos"
APP_DIR="$BASE_DIR/application"
BIN_DIR="$APP_DIR/bin"
BUILD_DIR="$BASE_DIR/tmp"
PKG_DIR="$BUILD_DIR/darwinpkg"
OUT_DIR="dist/macos"
VERSION=${VERSION:=dev}
FINAL_PACKAGE="$OUT_DIR/weep-installer-macos-$VERSION.pkg"

rm -rf "$BIN_DIR"
rm -rf "$BUILD_DIR"
mkdir -p "$BIN_DIR"
mkdir -p "$OUT_DIR"
mkdir -p "$PKG_DIR"

cp -r "$BASE_DIR/darwin" "$BUILD_DIR/"
chmod -R 755 "$BUILD_DIR/darwin/scripts"
chmod 755 "$BUILD_DIR/darwin/Distribution.xml"

printf "🟢 starting build for %s\n" "$FINAL_PACKAGE"

function prep_package() {
  # Prepare package structure
  mkdir -p "$BUILD_DIR/darwinpkg/Library/weep"
  cp -a "$APP_DIR/." "$BUILD_DIR/darwinpkg/Library/weep"
  chmod -R 755 "$BUILD_DIR/darwinpkg/Library/weep"

  # Replace tokens in package files
  sed -i '' -e "s/__VERSION__/$VERSION/g" ${BUILD_DIR}/darwin/Resources/*.html
}

function combine_binaries() {
  printf "🦾 creating universal binary..."
  output=$1
  bin1=$2
  bin2=$3
  lipo -create -output "$output" "$bin1" "$bin2"
  printf " done ✅ \n"
}

function sign_binary() {
  printf "🔏 signing binary..."
  binary=$1
  codesign \
    --options runtime \
    --sign "Developer ID Application: Netflix, Inc." \
    --force \
    --timestamp=http://timestamp.apple.com/ts01 \
    "$binary" > /dev/null 2>&1
  printf " done ✅ \n"
}

function build_package() {
  printf "📦 building package..."
  pkgbuild --identifier "com.netflix.weep" \
    --version "$VERSION" \
    --scripts "$BUILD_DIR/darwin/scripts" \
    --root "$BUILD_DIR/darwinpkg" \
    weep.pkg > /dev/null 2>&1

  productbuild --distribution "$BUILD_DIR/darwin/Distribution.xml" \
    --resources "$BUILD_DIR/darwin/Resources" \
    --package-path "$BUILD_DIR/package" \
    "$OUT_DIR/weep-$VERSION-unsigned.pkg" > /dev/null 2>&1
  printf " done ✅ \n"
}

function sign_package() {
  printf "🔏 signing package..."
  productsign --sign "Developer ID Installer: Netflix, Inc." \
    "$OUT_DIR/weep-$VERSION-unsigned.pkg" \
    "$FINAL_PACKAGE" > /dev/null 2>&1

  pkgutil --check-signature "$FINAL_PACKAGE" > /dev/null 2>&1
  printf " done ✅ \n"
}

function notarize() {
  printf "🔐 submitting package for notarization..."
  output=$(xcrun altool \
    --notarize-app \
    --primary-bundle-id "com.netflix.weep" \
    --username "psanders@netflix.com" \
    --password "$AC_PASSWORD" \
    --file "$FINAL_PACKAGE")
  printf " done ✅ \n"
  request_id=$(echo "$output" | grep RequestUUID | awk '{ print $3 }')
  printf "👨‍💻 waiting for Apple\n"
  printf "💡 notarize request id is %s\n" "$request_id"
  # give the server side a few seconds to sort things out
  sleep 5
  while true; do
    status=$(check_notarize_status "$request_id")
    printf "👀 current status \"%s\"" "$status"
    case "$status" in
      "success")
        printf ", done ✅ \n"
        break
        ;;
      "failure")
        printf ", exiting! 🔴 \n"
        exit 1
        ;;
      *)
        printf ", not ready yet 😴 \n"
        sleep 10
        ;;
    esac
  done
}

function check_notarize_status() {
  request_id=$1
  output=$(xcrun altool \
    --notarization-info "$request_id" \
    --username "psanders@netflix.com" \
    --password "$AC_PASSWORD")
  status=$(echo "$output" | grep "Status:" | awk '{ for (i=2; i<=NF; i++) printf("%s ", $i) }' | awk '{$1=$1;print}')
  echo "$status"
}

function staple() {
  printf "📎 stapling..."
  xcrun stapler staple "$FINAL_PACKAGE" > /dev/null 2>&1
  printf " done ✅ \n"
}

function cleanup() {
  rm dist/macos/*-unsigned.pkg
}

combine_binaries "$BIN_DIR/weep-universal" \
  dist/bin/weep_darwin_amd64/weep \
  dist/bin/weep_darwin_arm64/weep
sign_binary "$BIN_DIR/weep-universal"
prep_package
build_package
sign_package
notarize
staple
cleanup

printf "🙌 successfully built and notarized %s 🎉 \n" "$FINAL_PACKAGE"
