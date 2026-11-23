package GperlMod;

# The parts of dependency management that are Perl's own semantics —
# evaluating a cpanfile, comparing versions, reading distribution metadata —
# run here, on the interpreter's own CPAN::Meta toolchain. Every sub returns
# a flat list of strings.

use strict;
use warnings;
use CPAN::Meta;
use CPAN::Meta::Requirements;
use Module::Metadata;
use version ();

# cpanfile($path): (module, range, phase, relationship)* of every prereq the
# file declares outside feature blocks.
sub cpanfile {
    my ($path) = @_;
    open my $fh, '<', $path or die "$path: $!\n";
    my $src = do { local $/; <$fh> };
    close $fh;
    local @GperlMod::DSL::out   = ();
    local $GperlMod::DSL::phase = 'runtime';
    local $GperlMod::DSL::skip  = 0;
    eval "package GperlMod::DSL;\n#line 1 \"$path\"\n$src\n;1" or die $@;
    return @GperlMod::DSL::out;
}

# best($range, @versions): index of the highest version the range accepts,
# -1 when none does.
sub best {
    my ($range, @versions) = @_;
    my $req = CPAN::Meta::Requirements->from_string_hash({ m => $range });
    my ($best, $index) = (undef, -1);
    for my $i (0 .. $#versions) {
        my $text = $versions[$i] eq 'undef' ? 0 : $versions[$i];
        my $v = eval { version->parse($text) };
        next unless defined $v && $req->accepts_module(m => $text);
        ($best, $index) = ($v, $i) if !defined $best || $v > $best;
    }
    return $index;
}

# bundled($module): the version of the module this perl already has ("0"
# when it declares none), or the empty string when it has none. "perl"
# answers with the perl version itself.
sub bundled {
    my ($module) = @_;
    return "$]" if $module eq 'perl';
    my $meta = Module::Metadata->new_from_module($module) or return '';
    my $v = $meta->version;
    return defined $v ? "$v" : '0';
}

sub _load_meta {
    my ($dir) = @_;
    # gperl/MYMETA.json is what the distribution's configure step produced.
    for my $file (qw(gperl/MYMETA.json META.json META.yml)) {
        next unless -f "$dir/$file";
        my $meta = eval { CPAN::Meta->load_file("$dir/$file", { lazy_validation => 1 }) };
        return $meta if $meta;
    }
    return;
}

# meta($dir): what a distribution tree says about itself, as key/value
# records separated by the key "-":
#   dynamic_config 0|1, static_install 0|1, has_xs 0|1,
#   prereq module range phase relationship, provides package version
sub meta {
    my ($dir) = @_;
    my @out;
    my $meta = _load_meta($dir);
    my $has_xs = 0;
    require File::Find;
    File::Find::find({ no_chdir => 1, wanted => sub {
        $has_xs = 1 if /\.(?:xs|c|cc|cpp|cxx)\z/ && $File::Find::name !~ m{/(?:t|xt|inc|examples?|eg)/};
    } }, $dir);
    push @out, '-', 'has_xs', $has_xs;
    if ($meta) {
        push @out, '-', 'dynamic_config', ($meta->dynamic_config ? 1 : 0);
        push @out, '-', 'static_install', (($meta->custom('x_static_install') || 0) == 1 ? 1 : 0);
        my $prereqs = $meta->effective_prereqs;
        for my $phase (qw(configure build test runtime)) {
            for my $type (qw(requires recommends suggests)) {
                my $reqs = $prereqs->requirements_for($phase, $type)->as_string_hash;
                push @out, '-', 'prereq', $_, $reqs->{$_}, $phase, $type for sort keys %$reqs;
            }
        }
    } else {
        push @out, '-', 'dynamic_config', 1, '-', 'static_install', 0;
    }
    my $provides = $meta && $meta->provides;
    if ($provides && %$provides) {
        push @out, '-', 'provides', $_, (defined $provides->{$_}{version} ? "$provides->{$_}{version}" : 'undef')
            for sort keys %$provides;
    } elsif (-d "$dir/lib") {
        my $found = Module::Metadata->package_versions_from_directory("$dir/lib");
        my %no_index;
        if ($meta && (my $ni = $meta->no_index)) {
            $no_index{$_} = 1 for @{ $ni->{package} || [] };
        }
        for my $pkg (sort keys %$found) {
            next if $no_index{$pkg};
            my $v = $found->{$pkg}{version};
            push @out, '-', 'provides', $pkg, (defined $v ? "$v" : 'undef');
        }
    }
    return @out;
}

package GperlMod::DSL;

our (@out, $phase, $skip);

sub _prereq {
    my ($type, $module, @args) = @_;
    return if $skip;
    my ($version, %opts) = @args % 2 ? @args : (undef, @args);
    $version = $opts{version} unless defined $version;
    push @out, $module, (defined $version && length $version ? "$version" : '0'), $phase, $type;
}

sub requires   { _prereq('requires',   @_) }
sub recommends { _prereq('recommends', @_) }
sub suggests   { _prereq('suggests',   @_) }
sub conflicts  { _prereq('conflicts',  @_) }

sub on {
    my ($name, $code) = @_;
    local $phase = $name;
    $code->();
}

sub configure_requires { local $phase = 'configure'; requires(@_) }
sub build_requires     { local $phase = 'build';     requires(@_) }
sub test_requires      { local $phase = 'test';      requires(@_) }
sub author_requires    { local $phase = 'develop';   requires(@_) }

# Optional features are not selected.
sub feature {
    my $code = pop;
    local $skip = 1;
    $code->();
}

sub mirror { }
sub osname { }

1;
