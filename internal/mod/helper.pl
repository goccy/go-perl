package GperlMod;

# The parts of dependency management that are Perl's own semantics —
# evaluating a cpanfile, comparing versions, reading distribution metadata —
# run here, on the interpreter's own CPAN::Meta toolchain. Every sub returns
# a flat list of strings.

use strict;
use warnings;
use CPAN::Meta;
use CPAN::Meta::Requirements;
use File::Basename ();
use Module::Metadata;
use version ();

# cpanfile($path): every prereq the file declares, as (module, range, phase,
# relationship, feature, line, option count, (option, value)*)*; feature is
# empty outside feature blocks. The file runs in a package of its own, with
# warnings off, as Module::CPANfile runs it.
my $sandbox = 0;
sub cpanfile {
    my ($path) = @_;
    open my $fh, '<', $path or die "$path: $!\n";
    my $src = do { local $/; <$fh> };
    close $fh;
    local @GperlMod::DSL::out     = ();
    local $GperlMod::DSL::path    = $path;
    local $GperlMod::DSL::phase   = 'runtime';
    local $GperlMod::DSL::feature = '';
    my $pkg = 'GperlMod::Sandbox' . ++$sandbox;
    {
        no strict 'refs';
        *{"${pkg}::$_"} = \&{"GperlMod::DSL::$_"} for @GperlMod::DSL::bindings;
    }
    eval "package $pkg;\nno warnings;\n#line 1 \"$path\"\n$src\n;1" or die $@;
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

# The distributions Module::Build depends on (cpanm's should_use_mm).
my %use_mm = map { $_ => 1 } qw(version ExtUtils-ParseXS ExtUtils-Install ExtUtils-Manifest);

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
    # As cpanm does, a Build.PL distribution that requires nothing to
    # configure requires Module::Build, unless it is one Module::Build
    # itself needs (cpanm configures those with Makefile.PL).
    my ($name) = File::Basename::basename($dir) =~ /^([^\@]*)/;
    if (-f "$dir/Build.PL" && !$use_mm{$name}
        && !($meta && %{ $meta->effective_prereqs->requirements_for('configure', 'requires')->as_string_hash })) {
        push @out, '-', 'prereq', 'Module::Build', '0.38', 'configure', 'requires';
    }
    my $provides = $meta && $meta->provides;
    if ($provides && %$provides) {
        push @out, '-', 'provides', $_, (defined $provides->{$_}{version} ? "$provides->{$_}{version}" : 'undef')
            for sort keys %$provides;
    } elsif (-d "$dir/lib") {
        my $found = _lib_packages($dir);
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

# _lib_packages($dir): the packages under $dir/lib and their versions, read
# as the distribution's own build tools read them: from its directory, so a
# version line that loads a module finds it in lib/, and with a version line
# that cannot be evaluated giving an undefined version, as cpanm records it.
sub _lib_packages {
    my ($dir) = @_;
    my $evaluate = Module::Metadata->can('_evaluate_version_line')
        or die "Module::Metadata $Module::Metadata::VERSION has no _evaluate_version_line\n";
    require Cwd;
    my $cwd = Cwd::getcwd() // die "getcwd: $!\n";
    chdir $dir or die "chdir $dir: $!\n";
    my $found;
    my $ok = eval {
        local *Module::Metadata::_evaluate_version_line = sub {
            my @args = @_;
            my $v = eval { $evaluate->(@args) };
            warn $@ if $@;
            return $v;
        };
        $found = Module::Metadata->package_versions_from_directory("$dir/lib");
        1;
    };
    my $err = $@;
    chdir $cwd or die "chdir $cwd: $!\n";
    die $err unless $ok;
    return $found;
}

package GperlMod::DSL;

# The functions a cpanfile calls, bound into its package as
# Module::CPANfile::Environment binds them, and reading their arguments by
# the same rules.
our @bindings = qw(
    on requires recommends suggests conflicts feature osname mirror
    configure_requires build_requires test_requires author_requires
);

our (@out, $path, $phase, $feature);

# _line: the line of the cpanfile the running function was called from.
sub _line {
    for (my $i = 1; my @frame = caller($i); $i++) {
        return $frame[2] if $frame[1] eq $path;
    }
    return 0;
}

sub _fail { die "$path line " . _line() . ": $_[0]\n" }

sub _prereq {
    my ($type, $module, @args) = @_;
    _fail("$type needs a module name") if !defined $module || ref $module || $module eq '';
    my $version = @args % 2 ? shift @args : 0;
    my %opts = @args;
    $version = delete $opts{version} if exists $opts{version};
    $version ||= 0;
    for my $key (sort keys %opts) {
        _fail("$module: option $key needs a string") if !defined $opts{$key} || ref $opts{$key};
    }
    eval { CPAN::Meta::Requirements->new->add_string_requirement($module => "$version"); 1 }
        or _fail("$module: " . ($@ =~ s/\s+\z//r));
    push @out, $module, "$version", $phase, $type, $feature, _line(),
        scalar(keys %opts), map { ($_, "$opts{$_}") } sort keys %opts;
}

sub requires   { _prereq('requires',   @_) }
sub recommends { _prereq('recommends', @_) }
sub suggests   { _prereq('suggests',   @_) }
sub conflicts  { _prereq('conflicts',  @_) }

my %phases = map { $_ => 1 } qw(configure build test runtime develop);

sub on {
    my ($name, $code) = @_;
    _fail("usage: on 'phase' => sub { ... }") if !defined $name || ref $name || ref $code ne 'CODE';
    _fail("unknown phase '$name'") unless $phases{$name};
    local $phase = $name;
    $code->();
}

sub configure_requires { local $phase = 'configure'; requires(@_) }
sub build_requires     { local $phase = 'build';     requires(@_) }
sub test_requires      { local $phase = 'test';      requires(@_) }
sub author_requires    { local $phase = 'develop';   requires(@_) }

sub feature {
    my ($identifier, $description, $code) = @_;
    # feature 'identifier' => sub { ... }
    ($description, $code) = ($identifier, $description) if @_ == 2 && ref $description eq 'CODE';
    _fail("usage: feature 'identifier', 'Description' => sub { ... }")
        if !defined $identifier || ref $identifier || $identifier eq '' || ref $description ne '' || ref $code ne 'CODE';
    local $feature = $identifier;
    $code->();
}

sub osname { _fail('osname is not supported') }

sub mirror { _fail('the mirror statement is not supported; name a mirror with the dist and mirror options of a requirement') }

1;
